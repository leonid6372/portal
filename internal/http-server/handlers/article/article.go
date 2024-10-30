package article

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/mssql"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"portal/internal/storage/postgres/entities/user"
	"strconv"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	PostID int
}

type CommentInfo struct {
	news.Comment
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
}

type Article struct {
	Text     string        `json:"text"`
	Comments []CommentInfo `json:"comments"`
}

type Response struct {
	resp.Response
	Article Article `json:"article"`
}

func New(log *slog.Logger, storage *postgres.Storage, storage1C *mssql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.article.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		var err error

		// Считываем параметры запроса из request
		r.ParseForm()
		rawPostID, ok := r.Form["post_id"]
		if ok {
			req.PostID, err = strconv.Atoi(rawPostID[0])
			if err != nil {
				log.Error("failed to make int post id", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to make int post id"))
				return
			}
		} else {
			log.Error("empty post id parameter")
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("empty post id parameter"))
			return
		}

		// Получаем текст поста в p по ID поста
		var p news.Post
		if err := p.GetText(storage, req.PostID); err != nil {
			log.Error("failed to get post text", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get post text"))
			return
		}

		// Получаем все комментарии по post ID
		var c news.Comment
		cs, err := c.GetCommentsByPostID(storage, req.PostID)
		if err != nil {
			log.Error("failed to get comments", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get comments"))
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

		article := Article{
			Text:     p.Text,
			Comments: csi,
		}

		log.Info("article data successfully gotten")

		responseOK(w, r, log, article)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, article Article) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Article:  article,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
