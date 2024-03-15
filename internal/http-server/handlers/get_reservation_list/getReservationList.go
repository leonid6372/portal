package getReservationList

import (
	"net/http"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/Reservation"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
)

type Response struct {
	resp.Response
	ReservationList string `json:"reservation_list"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.getReservationList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var p *Reservation.Place
		placeList, err := p.GetActualPlaceList(storage)
		if err != nil {
			log.Error("failed to get reservation list")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))

			return
		}

		log.Info("actual reservation list gotten")

		responseOK(w, r, placeList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, shopList string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		ReservationList: reservationList,
	})
}
