package deletePost

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	minioServer "portal/internal/storage/minio"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"portal/internal/structs/roles"
	"slices"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	PostID int `json:"post_id" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage, miniosrv *minioServer.MinioProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.deleteArticle.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Определяем разрешенные роли
		allowedRoles := []int{roles.NewsEditor, roles.SuperAdmin}

		// Получаем user role из токена авторизации
		role := r.Context().Value(oauth.ScopeContext).(int)
		if role == 0 {
			log.Error("no user role in token")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user role in token"))
			return
		}

		//  Проверяем доступно ли действие для роли текущего пользователя
		if !slices.Contains(allowedRoles, role) {
			log.Error("access was denied")
			w.WriteHeader(403)
			render.JSON(w, r, resp.Error("access was denied"))
			return
		}

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

		// Удаляем новость из БД
		var p news.Post
		if err := p.DeletePost(storage, req.PostID); err != nil {
			log.Error("failed to delete post", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to delete post"))
			return
		}

		log.Info("post successfully deleted")

		render.JSON(w, r, resp.OK())
	}
}
