package lockerReservation

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	LockerID int `json:"locker_id" validate:"required"`
	Start    int `json:"start" validate:"required"`
	Finish   int `json:"finish" validate:"required"`
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

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		req.Start /= 1000 // Cut three time zone zeroes at the end
		rawStart := time.Unix(int64(req.Start), 0)
		start := rawStart.Format(time.DateOnly)

		req.Finish /= 1000 // Cut three time zone zeroes at the end
		rawFinish := time.Unix(int64(req.Finish), 0)
		finish := rawFinish.Format(time.DateOnly) + " 23:59:00"

		// Проверка наличия брони у пользователя в эту дату
		var locker_reservation *reservation.LockerReservation
		hasUserLockerReservation, err := locker_reservation.HasUserLockerReservationInDateRange(storage, userID, start, finish)
		if err != nil {
			log.Error("failed to check has user locker reservation if date range", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to check has user locker reservation if date range"))
			return
		}
		if hasUserLockerReservation {
			log.Error(fmt.Sprintf("%s: user already has locker reservation in date range"), op)
			w.WriteHeader(406)
			render.JSON(w, r, resp.Error("user already has locker reservation in date range"))
			return
		}

		// Добавление записи бронирования в БД
		err = locker_reservation.InsertLockerReservation(storage, req.LockerID, userID, start, finish)
		if err != nil {
			log.Error("failed to reserve locker", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to reserve locker"))
			return
		}

		log.Info("locker successfully reserved")

		render.JSON(w, r, resp.OK())
	}
}
