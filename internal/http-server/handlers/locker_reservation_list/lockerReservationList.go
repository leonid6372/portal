package lockerReservationList

import (
	"encoding/json"
	"net/http"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"
	"portal/internal/storage/postgres/entities/user"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
)

type Request struct {
	Start  time.Time `json:"start"`
	Finish time.Time `json:"finish"`
}

type ActualLockerInfo struct {
	reservation.ActualLocker
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
}

type Response struct {
	resp.Response
	ActualLockers []ActualLockerInfo `json:"locker_reservation_list"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.lockerReservationList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Декодируем json запроса
		r.ParseForm()
		rawStart, ok := r.Form["start"]
		if ok {
			intStart, err := strconv.Atoi(rawStart[0][:len(rawStart[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert locker reservation start time", sl.Err(err))
			}
			req.Start = time.Unix(int64(intStart), 0)
		}

		rawFinish, ok := r.Form["finish"]
		if ok {
			intFinish, err := strconv.Atoi(rawFinish[0][:len(rawFinish[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert locker reservation finish time", sl.Err(err))
			}
			req.Finish = time.Unix(int64(intFinish), 0)
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Проверям были ли заданы параметры начала и конца бронирования
		if req.Start.IsZero() && req.Finish.IsZero() {
			req.Start = time.Now()
			req.Finish = time.Now().Add(time.Hour)
		}

		// Запрашиваем свободные места согласно заданным параметарм бронирования
		var al reservation.ActualLocker
		als, err := al.GetActualLockers(storage, req.Start, req.Finish)
		if err != nil {
			log.Error("failed to get locker reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get locker reservation list"))
			return
		}

		// Подготовливаем итоговую структуру со всей информацией о бронированиях
		var alsi []ActualLockerInfo
		for _, al := range als {
			ali := ActualLockerInfo{ActualLocker: al}
			// Если место занято кем-то, то запрашиваем в 1С его ФИО и добавляем к инфо о брони
			if ali.UserID != 0 {
				var u user.User
				err := u.GetUsername(storage, ali.UserID)
				if err != nil {
					log.Error("failed to get username", sl.Err(err))
					w.WriteHeader(422)
					render.JSON(w, r, resp.Error("failed to get username"))
					return
				}
				err = u.GetUserInfo(storage, u.Username)
				if err != nil {
					log.Error("failed to get user info", sl.Err(err))
					w.WriteHeader(422)
					render.JSON(w, r, resp.Error("failed to get user info"))
					return
				}

				ali.FullName = u.FullName
				ali.Position = u.Position
				ali.Department = u.Department
			}
			alsi = append(alsi, ali)
		}

		log.Info("actual locker reservation list gotten")

		responseOK(w, r, log, alsi)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, actualLockers []ActualLockerInfo) {
	response, err := json.Marshal(Response{
		Response:      resp.OK(),
		ActualLockers: actualLockers,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
