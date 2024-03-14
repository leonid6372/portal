package reservation

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/Reservation"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	PlaceID int    `json:"place_id,omitempty" validate:"required"`
	Start   string `json:"start,omitempty" validate:"required"`
	Finish  string `json:"finish,omitempty" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.reservation.New"

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

		var res *Reservation.Reservation
		status, err := res.ReservationInsert(storage, req.PlaceID, req.Start, req.Finish)
		if !status {
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		// Обработка недоступности указанного place_id

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
