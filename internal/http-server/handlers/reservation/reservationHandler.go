package reservationHandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	PlaceID int    `json:"place_id" validate:"required"`
	Start   string `json:"start" validate:"required"`
	Finish  string `json:"finish" validate:"required"`
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
			render.JSON(w, r, resp.Error("empty request: "+err.Error()))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("failed to decode request: "+err.Error()))
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
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]string)
		userID, err := strconv.Atoi(tempUserID["user_id"])
		if err != nil {
			log.Error("failed to get user id from token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to get user id from token claims: "+err.Error()))
			return
		}

		// Добавление записи бронирования в БД
		var reservation *reservation.Reservation
		err = reservation.InsertReservation(storage, req.PlaceID, userID, req.Start, req.Finish)
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to reserve place: "+err.Error()))
			return
		}

		log.Info("place successfully reserved")

		render.JSON(w, r, resp.OK())
	}
}
