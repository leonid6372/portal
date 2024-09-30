package profile

import (
	"encoding/json"
	"fmt"
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

const (
	qrGetUserByUsername = `SELECT _Fld7252, _Fld7254, _Fld7255, _Fld7256, _Fld7257 FROM [10295].[dbo].[_InfoRg7251] WHERE _Fld7252 = $1;`
)

// Временная вспомогательная структура
type Profile struct {
	Fld7252 string `json:"username"`   // username
	Fld7254 string `json:"full_name"`  // fullname
	Fld7255 string `json:"position"`   // position
	Fld7256 string `json:"department"` // department
	Fld7257 string `json:"birthday"`   // birthday
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

		// Получаем данные пользователя по username из БД MSSQL
		p := Profile{Fld7252: u.Username}
		err = p.GetUserByUsername(storage1C)
		if err != nil {
			log.Error("failed to get profile", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get profile"))
			return
		}

		log.Info("profile data successfully gotten")

		responseOK(w, r, log, p)
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

func (p *Profile) GetUserByUsername(storage1C *mssql.Storage) error {
	const op = "storage.postgres.entities.user.GetUserByUsername"

	// Проверяем username в БД 1С
	stmt, err := storage1C.DB.Prepare(qrGetUserByUsername)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	qrResult, err := stmt.Query(p.Fld7252)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return fmt.Errorf("%s: wrong username", op)
	}

	if err := qrResult.Scan(&p.Fld7252, &p.Fld7254, &p.Fld7255, &p.Fld7256, &p.Fld7257); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
