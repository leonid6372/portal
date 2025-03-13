package me

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type User struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      int    `json:"role"`
	ImagePath string `json:"image_path"`
}

type Response struct {
	resp.Response
	User User `json:"user"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.me.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		// Получаем username из токена авторизации
		username := r.Context().Value(oauth.CredentialContext).(string)
		if username == "" {
			log.Error("no username in token")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no username in token"))
			return
		}

		// Получаем роль из токена авторизации
		userRole := r.Context().Value(oauth.ScopeContext).(int)
		if userRole == 0 {
			log.Error("no user role in token")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user role in token"))
			return
		}

		// Получаем путь фото пользователя из БД
		var u user.User
		err := u.GetImagePath(storage, userID)
		if err != nil {
			log.Error("failed to get image path", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get image path"))
			return
		}

		user := User{UserID: userID, Username: username, Role: userRole, ImagePath: u.ImagePath}

		responseOK(w, r, log, user)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, user User) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		User:     user,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
