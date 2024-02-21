package logIn

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/jwt"
	"portal/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Login    string `json:"login,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type Response struct {
	resp.Response
	Token string `json:"token"`
}

type UserGetter interface {
	GetUser(login, password string) (bool, error)
}

func New(log *slog.Logger, userGetter UserGetter, tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.logIn.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Декодируем json запроса
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
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

			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// Проверка данных пользователя по БД
		//ok, err := userGetter.GetUser(req.Login, req.Password)
		/*if err != nil {

		}*/

		// Создание токена и отправка полем JSON в ответ запроса
		token, _ := jwt.New(tokenAuth)
		responseOK(w, r, token)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, token string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Token:    token,
	})
}
