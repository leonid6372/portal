package article

import (
	"encoding/json"
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
	PostID int `json:"post_id" validate:"required"`
}

type Article struct {
	Text     string
	Comments []news.Comment
}

type Response struct {
	resp.Response
	Article Article
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.article.New"

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
			render.JSON(w, r, resp.Error("failed to decode request: "+err.Error()))
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

		// Получаем текст поста в p по ID поста
		var p news.Post
		if err := p.GetText(storage, req.PostID); err != nil {
			log.Error("failed to get post text", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get post text"))
			return
		}

		// Получаем все комментарии по post ID
		var c news.Comment
		cs, err := c.GetComments(storage, req.PostID)
		if err != nil {
			log.Error("failed to get post text", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get post text"))
			return
		}

		article := Article{
			Text:     p.Text,
			Comments: cs,
		}

		log.Info("article data successfully gotten")

		responseOK(w, r, log, article)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, article Article) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Article:  article,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
