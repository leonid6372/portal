package userReservations

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/reservation"
	"strconv"

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

		// Поулчаем user id из токена
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]string)
		userID, err := strconv.Atoi(tempUserID["user_id"])
		if err != nil {
			log.Error("failed to get user id from token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to get user id from token claims: "+err.Error()))
			return
		}

		var reservations reservation.Reservations
		if err := reservations.GetReservationsByUserID(storage, userID); err != nil {
			log.Error("failed to get reservation list")
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list: "+err.Error()))
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
		log.Error("failed to process response")
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response: "+err.Error()))
		return
	}

	render.Data(w, r, response)
}
