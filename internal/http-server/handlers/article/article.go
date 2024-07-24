package article

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"strconv"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	PostID int
}

type Article struct {
	Text     string         `json:"text"`
	Comments []news.Comment `json:"comments"`
}

type Response struct {
	resp.Response
	Article Article `json:"article"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.article.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		var err error

		// Считываем параметры запроса из request
		r.ParseForm()
		rawPostID, ok := r.Form["post_id"]
		if ok {
			req.PostID, err = strconv.Atoi(rawPostID[0])
			if err != nil {
				log.Error("failed to make int post id", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to make int post id"))
				return
			}
		} else {
			log.Error("empty post id parameter")
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("empty post id parameter"))
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
