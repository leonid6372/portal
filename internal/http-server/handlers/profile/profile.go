package profile

import (
	"encoding/json"
	"net/http"

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

type Response struct {
	resp.Response
	Profile Data1C `json:"profile"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.profile.New"

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

		// Получаем данные пользователя по user id из БД
		u := user.User{UserID: userID}
		err := u.GetUserById(storage)
		if err != nil {
			log.Error("failed to get profile")
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get profile: "+err.Error()))
			return
		}

		log.Info("profile data successfully gotten")

		// Декодируем полученный из БД JSON с данными профиля
		var data1C Data1C
		if err = json.Unmarshal([]byte(u.Data1C), &data1C); err != nil {
			log.Error("failed to process response")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to process response: "+err.Error()))
			return
		}

		responseOK(w, r, log, data1C)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, profile Data1C) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Profile:  profile,
	})
	if err != nil {
		log.Error("failed to process response")
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response: "+err.Error()))
		return
	}

	render.Data(w, r, response)
}
