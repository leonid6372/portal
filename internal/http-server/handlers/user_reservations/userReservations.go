package userReservations

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/reservation"
	"portal/internal/structs/roles"
	"slices"

	"portal/internal/lib/oauth"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type ReservationInfo struct {
	reservation.Reservation
	PlaceName string `json:"place_name"`
}

type Response struct {
	resp.Response
	ReservationsInfo []ReservationInfo `json:"user_reservations"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.userReservations.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Определяем разрешенные роли
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

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		var reserv reservation.Reservation
		var reservations []reservation.Reservation
		reservations, err := reserv.GetReservationsByUserID(storage, userID)
		if err != nil {
			log.Error("failed to get reservation list", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get reservation list"))
			return
		}

		// Добавляем имя места к списку бронирований
		var rsi []ReservationInfo
		for _, reserv := range reservations {
			ri := ReservationInfo{Reservation: reserv}

			var place reservation.Place
			err = place.GetPlaceName(storage, reserv.PlaceID)
			if err != nil {
				log.Error("failed to get place name", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get place name"))
				return
			}

			ri.PlaceName = place.Name
			rsi = append(rsi, ri)
		}

		log.Info("user reservations gotten")

		responseOK(w, r, log, rsi)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, reservationsInfo []ReservationInfo) {
	response, err := json.Marshal(Response{
		Response:         resp.OK(),
		ReservationsInfo: reservationsInfo,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
