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
	ShopList []shop.Item `json:"shopList"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.shopList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var i *shop.Item
		rawShopList, err := i.GetShopList(storage)
		if err != nil {
			log.Error("failed to get shop list")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get shop list"))

			return
		}

		log.Info("shop list gotten")

		var shopList []shop.Item
		if err = json.Unmarshal([]byte(rawShopList), &shopList); err != nil {
			log.Error("failed to process response")

			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to process response"))

			return
		}

		responseOK(w, r, log, shopList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, shopList []shop.Item) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		ShopList: shopList,
	})
	if err != nil {
		log.Error("failed to process response")

		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))

		return
	}

	render.Data(w, r, response)
}
