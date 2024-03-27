package order

import (
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.updateCartItem.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var c *shop.Cart
		err := c.Order(storage, 1)

		// Обработка общего случая ошибки БД
		if err != nil {
			log.Error(err.Error())

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to complete order sequence"))

			return
		}

		log.Info("order successfully made")

		render.JSON(w, r, resp.OK())
	}
}
