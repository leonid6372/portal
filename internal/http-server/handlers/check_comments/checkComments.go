package checkComments

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/mssql"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"portal/internal/storage/postgres/entities/user"
	"portal/internal/structs/roles"
	"slices"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type CommentInfo struct {
	news.Comment
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
}

type Response struct {
	resp.Response
	Comments []CommentInfo `json:"comments"`
}

func New(log *slog.Logger, storage *postgres.Storage, storage1C *mssql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.checkComments.New"

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

		// Получаем все комментарии по post ID
		var c news.Comment
		cs, err := c.GetUncheckedComments(storage)
		if err != nil {
			log.Error("failed to get unchecked comments", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get unchecked comments"))
			return
		}

		// Подготовливаем итоговую структуру со всей информацией комментариях в посте
		var csi []CommentInfo
		for _, c := range cs {
			ci := CommentInfo{Comment: c}
			// Запрашиваем ФИО пользователя, оставившего комментарий к посту
			var u user.User
			err := u.GetUsername(storage, ci.UserID)
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

			ci.FullName = u.FullName
			ci.Position = u.Position
			ci.Department = u.Department
			csi = append(csi, ci)
		}

		responseOK(w, r, log, csi)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, commentsInfo []CommentInfo) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Comments: commentsInfo,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
