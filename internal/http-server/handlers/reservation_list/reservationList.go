package reservationList

import (
	"encoding/json"
	"strconv"
	"net/http"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
)

type Request struct {
	Start      time.Time `json:"start"`
	Finish     time.Time `json:"finish"`
	Properties string    `json:"properties"`
}

type Response struct {
	resp.Response
	ActualPlaces []reservation.ActualPlace `json:"reservation_list"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.reservationList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Декодируем json запроса
		r.ParseForm()
		intStart, err := strconv.Atoi(r.Form["start"][0][:len(r.Form["start"][0])-3]) // Cut three time zone zeroes at the end
		if err != nil {
			log.Error("failed to convert reservation start time", sl.Err(err))
		}
		req.Start = time.Unix(int64(intStart), 0)

		intFinish, err := strconv.Atoi(r.Form["finish"][0][:len(r.Form["finish"][0])-3]) // Cut three time zone zeroes at the end
		if err != nil {
			log.Error("failed to convert reservation finish time", sl.Err(err))
		}
		req.Finish = time.Unix(int64(intFinish), 0)

		log.Info("request body decoded", slog.Any("request", req))

		// Проверям были ли заданы параметры начала и конца бронирования
		if req.Start.IsZero() && req.Finish.IsZero() {
			req.Start = time.Now()
			req.Finish = time.Now().Add(time.Hour)
		}

		// Запрашиваем свободные места согласно заданным параметарм бронирования
		var ap reservation.ActualPlace
		// TO DO: проработать поиск по параметрам рабочего места
		aps, err := ap.GetActualPlaces(storage, req.Properties, req.Start, req.Finish)
		if err != nil {
			log.Error("failed to get reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))
			return
		}

		log.Info("actual reservation list gotten")

		responseOK(w, r, log, aps)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, actualPlaces []reservation.ActualPlace) {
	response, err := json.Marshal(Response{
		Response:     resp.OK(),
		ActualPlaces: actualPlaces,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
