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
	UserReservations []reservation.Reservation `json:"userReservations"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.userReservations.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]string)
		userID, err := strconv.Atoi(tempUserID["user_id"])
		if err != nil {
			log.Error("failed to get user id from token claimss")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to get user id from token claims"))
			return
		}

		var rsrv *reservation.Reservation
		rawUserReservations, err := rsrv.GetUserReservations(storage, userID)
		if err != nil {
			log.Error("failed to get user reservations")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get user reservations"))

			return
		}

		log.Info("user reservations gotten")

		var userReservations []reservation.Reservation
		if err = json.Unmarshal([]byte(rawUserReservations), &userReservations); err != nil {
			log.Error("failed to process response")

			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to process response"))

			return
		}

		responseOK(w, r, log, userReservations)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, userReservations []reservation.Reservation) {
	response, err := json.Marshal(Response{
		Response:         resp.OK(),
		UserReservations: userReservations,
	})
	if err != nil {
		log.Error("failed to process response")

		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))

		return
	}

	render.Data(w, r, response)
}
