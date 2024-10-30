package profile

import (
	"encoding/json"
	"net/http"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/mssql"
)

// Временная вспомогательная структура
type Profile struct {
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
	//Birthday   string `json:"birthday"`
}

type Response struct {
	resp.Response
	Profile Profile `json:"profile"`
}

func New(log *slog.Logger, storage *postgres.Storage, storage1C *mssql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.profile.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var err error

		var userID int

		// Считываем параметры запроса из request
		r.ParseForm()
		rawUserID, ok := r.Form["user_id"]
		if ok {
			userID, err = strconv.Atoi(rawUserID[0])
			if err != nil {
				log.Error("failed to make int user id", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to make int user id"))
				return
			}
		} else {
			// Получаем userID из токена авторизации, если не указан в параметре
			tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
			userID, ok = tempUserID["user_id"]
			if !ok {
				log.Error("no user id in token claims")
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("no user id in token claims"))
				return
			}
		}

		// Получаем username из БД
		var u user.User
		err = u.GetUsername(storage, userID)
		if err != nil {
			log.Error("failed to get username", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get username"))
			return
		}

		// Получаем user info из БД
		err = u.GetUserInfo(storage, u.Username)
		if err != nil {
			log.Error("failed to get user info", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get user info"))
			return
		}

		log.Info("profile data successfully gotten")

		responseOK(w, r, log, Profile{Username: u.Username, FullName: u.FullName, Position: u.Position, Department: u.Department})
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, profile Profile) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Profile:  profile,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
