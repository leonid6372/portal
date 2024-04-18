package reservationList

import (
	"encoding/json"
	"errors"
	"io"
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
			render.JSON(w, r, resp.Error("failed to decode request: "+err.Error()))
			return
		}

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
			render.JSON(w, r, resp.Error("failed to get reservation list: "+err.Error()))
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
		render.JSON(w, r, resp.Error("failed to process response: "+err.Error()))
		return
	}

	render.Data(w, r, response)
}
