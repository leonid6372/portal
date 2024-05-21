package userReservations

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/reservation"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/oauth"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Reservations reservation.Reservations `json:"user_reservations"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.userReservations.New"

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

		var reservations reservation.Reservations
		if err := reservations.GetReservationsByUserID(storage, userID); err != nil {
			log.Error("failed to get reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))
			return
		}

		log.Info("user reservations gotten")

		responseOK(w, r, log, reservations)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, reservations reservation.Reservations) {
	response, err := json.Marshal(Response{
		Response:     resp.OK(),
		Reservations: reservations,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
