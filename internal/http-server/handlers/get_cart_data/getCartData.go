package get_cart_data

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
)

type Response struct {
	resp.Response
	CartData []CartData `json:"cart_data"`
}

type CartData struct {
	ItemID       int `json:"item_id"`
	Quantity     int `json:"quantity"`
	InCartItemID int `json:"in_cart_item_id"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.getCartData.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var c *shop.Cart
		cartData, err := c.GetCartData(storage, 1)
		if err != nil {
			log.Error("failed to get cart data")
			fmt.Printf("%s", err)

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get cart data"))

			return
		}

		log.Info("cart data loaded")

		var itemList []CartData
		_ = json.Unmarshal(cartData, &itemList)

		responseOK(w, r, itemList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, cartData []CartData) {
	resp, _ := json.Marshal(Response{
		Response: resp.OK(),
		CartData: cartData,
	})

	render.Data(w, r, resp)
}
