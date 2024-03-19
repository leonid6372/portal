package getShopList

import (
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
	ShopList string `json:"shop_list"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.getShopList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var i *shop.Item
		shopList, err := i.GetShopList(storage)
		if err != nil {
			log.Error("failed to get shop list")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get shop list"))

			return
		}

		log.Info("shop list gotten")

		responseOK(w, r, shopList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, shopList string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		ShopList: shopList,
	})
}
