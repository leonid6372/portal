package reservationList

import (
	"encoding/json"
	"net/http"
	"portal/internal/storage/mssql"
	"portal/internal/storage/postgres"
	reservation "portal/internal/storage/postgres/entities/reservation"
	"portal/internal/storage/postgres/entities/user"
	"portal/internal/structs/roles"
	"slices"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
)

type Request struct {
	Start      time.Time `json:"start"`
	Finish     time.Time `json:"finish"`
	Properties string    `json:"properties"`
}

type ActualPlaceInfo struct {
	reservation.ActualPlace
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
}

type Response struct {
	resp.Response
	ActualPlaces []ActualPlaceInfo `json:"reservation_list"`
}

func New(log *slog.Logger, storage *postgres.Storage, storage1C *mssql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.reservationList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Определяем запрещенные роли
		restrictedRoles := []int{roles.UserWithOutReservation}

		// Получаем user role из токена авторизации
		role := r.Context().Value(oauth.ScopeContext).(int)
		if role == 0 {
			log.Error("no user role in token")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user role in token"))
			return
		}

		//  Проверяем доступно ли действие для роли текущего пользователя
		if slices.Contains(restrictedRoles, role) {
			log.Error("access was denied")
			w.WriteHeader(403)
			render.JSON(w, r, resp.Error("access was denied"))
			return
		}

		var req Request

		// Декодируем json запроса
		r.ParseForm()
		rawStart, ok := r.Form["start"]
		if ok {
			intStart, err := strconv.Atoi(rawStart[0][:len(rawStart[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert reservation start time", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to convert reservation start time"))
				return
			}
			req.Start = time.Unix(int64(intStart), 0)
		}

		rawFinish, ok := r.Form["finish"]
		if ok {
			intFinish, err := strconv.Atoi(rawFinish[0][:len(rawFinish[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert reservation finish time", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to convert reservation start time"))
				return
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
		var ap reservation.ActualPlace
		// TO DO: проработать поиск по параметрам рабочего места
		aps, err := ap.GetActualPlaces(storage, req.Properties, req.Start, req.Finish)
		if err != nil {
			log.Error("failed to get reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))
			return
		}

		// Подготовливаем итоговую структуру со всей информацией о бронированиях
		var apsi []ActualPlaceInfo
		for _, ap := range aps {
			api := ActualPlaceInfo{ActualPlace: ap}
			// Если место занято кем-то, то запрашиваем в 1С его ФИО и добавляем к инфо о брони
			if api.UserID != 0 {
				var u user.User
				err := u.GetUsername(storage, api.UserID)
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

				api.FullName = u.FullName
				api.Position = u.Position
				api.Department = u.Department
			}
			apsi = append(apsi, api)
		}

		log.Info("actual reservation list gotten")

		responseOK(w, r, log, apsi)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, actualPlaces []ActualPlaceInfo) {
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
