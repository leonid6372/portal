package comment

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	PostID int    `json:"post_id" validate:"required"`
	Text   string `json:"text" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.comment.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Декодируем json запроса
		err := render.DecodeJSON(r.Body, &req)
		// Такую ошибку встретим, если получили запрос с пустым телом.
		// Обработаем её отдельно
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Валидация обязательных полей запроса
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		// Обновляем значение текста комментария в БД
		var c news.Comment
		err = c.NewComment(storage, req.Text, userID, req.PostID)
		if err != nil {
			log.Error("failed to add comment", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to add comment"))
			return
		}

		log.Info("comment successfully added")

		render.JSON(w, r, resp.OK())
	}
}
