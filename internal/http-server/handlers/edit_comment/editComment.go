package editComment

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	CommentID int    `json:"comment_id" validate:"required"`
	Text      string `json:"text" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.editComment.New"

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

		// Обновляем значение текста комментария в БД
		var c news.Comment
		err = c.UpdateCommentText(storage, req.CommentID, req.Text)
		if err != nil {
			log.Error("failed to update comment text", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to update comment text"))
			return
		}

		log.Info("comment successfully updated")

		render.JSON(w, r, resp.OK())
	}
}
