package dropCartItem

import (
	"errors"
	"io"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	InCartItemID int `json:"in_cart_item_id" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.dropCartItem.New"

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
			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Удаление item из корзины
		var ici *shop.InCartItem
		if err := ici.DeleteInCartItem(storage, req.InCartItemID); err != nil {
			log.Error("failed to delete in cart item", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to delete in cart item"))
			return
		}

		log.Info("item successfully deleted")

		render.JSON(w, r, resp.OK())
	}
}
