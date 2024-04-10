package shopList

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"
)

type Response struct {
	resp.Response
	Items []shop.Item `json:"shop_list"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.shopList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем слайс товаров
		var is shop.Items
		if err := is.GetItems(storage); err != nil {
			log.Error("failed to get shop list", err)
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get shop list: "+err.Error()))
			return
		}

		log.Info("shop list gotten")

		responseOK(w, r, log, is)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, items shop.Items) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Items:    items,
	})
	if err != nil {
		log.Error("failed to process response")
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response: "+err.Error()))
		return
	}

	render.Data(w, r, response)
}
