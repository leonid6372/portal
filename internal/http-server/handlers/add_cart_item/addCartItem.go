package addCartItem

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"

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

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Валидация обязательных полей запроса
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}
		fmt.Println("kkk")
		return
	}
}
