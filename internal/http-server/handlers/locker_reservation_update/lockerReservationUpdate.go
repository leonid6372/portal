package lockerReservationUpdate

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/reservation"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	LockerReservationID int       `json:"locker_reservation_id" validate:"required"`
	LockerID            int       `json:"locker_id" validate:"required"`
	Start               time.Time `json:"start" validate:"required"`
	Finish              time.Time `json:"finish" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.lockerReservationUpdate.New"

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

		// TO DO: нужна ли проверка, что обновляется свое бронирование
		// Обновление записи бронирования в БД
		var lockerReservation *reservation.LockerReservation
		err = lockerReservation.UpdateLockerReservation(storage, req.LockerReservationID, req.LockerID, req.Start, req.Finish)
		if err != nil {
			log.Error("failed to update locker reservation", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to update locker reservation"))
			return
		}

		log.Info("locker reservation successfully updated")

		render.JSON(w, r, resp.OK())
	}
}
