package addCartItem

import (
	"errors"
	"io"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	ItemID   int `json:"item_id,omitempty" validate:"required"`
	Quantity int `json:"quantity,omitempty" validate:"required"`
}

type Response struct {
	resp.Response
}

type CartItemAdder interface {
	AddCartItem(item_id, quantity int) error
}

func New(log *slog.Logger, cartItemAdder CartItemAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.addCartItem.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Декодируем json запроса
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
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

		err = cartItemAdder.AddCartItem(req.ItemID, req.Quantity)

		// Обработка недоступности указанного item_id для заказа
		if errors.As(err, &storage.ErrItemUnavailable) {
			log.Error(err.Error())

			w.WriteHeader(406)
			render.JSON(w, r, resp.Error("item is not available"))

			return
		}

		// Обработка общего случая ошибки БД
		if err != nil {
			log.Error(err.Error())

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to add item in cart"))

			return
		}

		log.Info("item successfully added")

		render.JSON(w, r, resp.OK())
	}
}
