package reservationList

import (
	"encoding/json"
	"net/http"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
)

type Response struct {
	resp.Response
	ReservationList []reservation.Place `json:"reservationList"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.reservationList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var p *reservation.Place
		rawPlaceList, err := p.GetActualPlaceList(storage)
		if err != nil {
			log.Error("failed to get reservation list")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))

			return
		}

		log.Info("actual reservation list gotten")

		var placeList []reservation.Place
		_ = json.Unmarshal([]byte(rawPlaceList), &placeList)

		responseOK(w, r, placeList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, reservationList []reservation.Place) {
	resp, _ := json.Marshal(Response{
		Response:        resp.OK(),
		ReservationList: reservationList,
	})

	render.Data(w, r, resp)
}
