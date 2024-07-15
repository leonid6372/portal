package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/mssql"
)

const (
	qrGetUserByUsername = `SELECT _Fld7252, _Fld7254, _Fld7255, _Fld7256, _Fld7257 FROM [10295].[dbo].[_InfoRg7251] WHERE _Fld7252 = $1;`
)

type Request struct {
	Username string `json:"username" validate:"required"`
}

// Временная вспомогательная структура
type Profile struct {
	Fld7252 string `json:"username"`   // username
	Fld7254 string `json:"fullname"`   // fullname
	Fld7255 string `json:"position"`   // position
	Fld7256 string `json:"department"` // department
	Fld7257 string `json:"birthday"`   // birthday
}

type Response struct {
	resp.Response
	Profile Profile `json:"profile"`
}

func New(log *slog.Logger, storage1C *mssql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.profile.New"

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
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Валидация обязательных полей запроса
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Получаем данные пользователя по username из БД MSSQL
		p := Profile{Fld7252: req.Username}
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
