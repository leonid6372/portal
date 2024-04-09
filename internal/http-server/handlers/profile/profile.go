package profile

import (
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"
)

// Временная вспомогательная структура
type Data1C struct {
	Position  string `json:"position"`
	FullName  string `json:"full_name"`
	PhotoPath string `json:"photo_path"`
}

// Временная вспомогательная структура
type UserInfo struct {
	Balance int    `json:"balance"`
	Data    Data1C `json:"data"`
}

type Response struct {
	resp.Response
	Profile UserInfo `json:"profile"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.profile.New"

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

		var u user.User
		u.UserID = userID
		err = u.GetUserById(storage)
		if err != nil {
			log.Error("failed to get user info")

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get user info"))

			return
		}

		log.Info("shop list gotten")

		var data1C Data1C
		if err = json.Unmarshal([]byte(u.Data1C), &data1C); err != nil {
			log.Error("failed to process response")

			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to process response"))

			return
		}

		userInfo := UserInfo{Balance: u.Balance, Data: data1C}

		responseOK(w, r, log, userInfo)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, userInfo UserInfo) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Profile:  userInfo,
	})
	if err != nil {
		log.Error("failed to process response")

		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))

		return
	}

	render.Data(w, r, response)
}
