package tags

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Tags []news.Tag `json:"tags"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.tags.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var tag news.Tag
		var tags []news.Tag
		tags, err := tag.GetTags(storage)
		if err != nil {
			log.Error("failed to get tags", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get tags"))
			return
		}

		log.Info("tags gotten")

		responseOK(w, r, log, tags)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, tags []news.Tag) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Tags:     tags,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
