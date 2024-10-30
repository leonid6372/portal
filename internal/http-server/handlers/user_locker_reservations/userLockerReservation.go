package userLockerReservations

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/reservation"

	"portal/internal/lib/oauth"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type LockerReservationInfo struct {
	reservation.LockerReservation
	LockerName string `json:"locker_name"`
}

type Response struct {
	resp.Response
	LockerReservationsInfo []LockerReservationInfo `json:"user_locker_reservations"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.userLockerReservations.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		var lockerReserv reservation.LockerReservation
		var lockerReservations []reservation.LockerReservation
		lockerReservations, err := lockerReserv.GetLockerReservationsByUserID(storage, userID)
		if err != nil {
			log.Error("failed to get locker reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get locker reservation list"))
			return
		}

		// Добавляем имя шкафчика к списку бронирований
		var lrsi []LockerReservationInfo
		for _, lockerReserv := range lockerReservations {
			lri := LockerReservationInfo{LockerReservation: lockerReserv}

			var locker reservation.Locker
			err = locker.GetLockerName(storage, lockerReserv.LockerID)
			if err != nil {
				log.Error("failed to get locker name", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get locker name"))
				return
			}

			lri.LockerName = locker.Name
			lrsi = append(lrsi, lri)
		}

		log.Info("user locker reservations gotten")

		responseOK(w, r, log, lrsi)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, lockerReservationsInfo []LockerReservationInfo) {
	response, err := json.Marshal(Response{
		Response:               resp.OK(),
		LockerReservationsInfo: lockerReservationsInfo,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
