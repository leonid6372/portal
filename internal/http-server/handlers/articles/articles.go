package articles

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	storageHandler "portal/internal/storage"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"strconv"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

var (
	errPageInOutOfRange = errors.New("page in out of range")
)

type Request struct {
	Tags []string
	Page int
}

// Запрашиваемая API структура
type Article struct {
	news.Post
	LikesAmount int        `json:"likes_amount"`
	Images      []string   `json:"images"`
	Tags        []news.Tag `json:"tags"`
}

type Response struct {
	resp.Response
	Articles []Article `json:"articles"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.articles.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		var err error
		var ok bool

		r.ParseForm()
		req.Tags = r.Form["tag"]
		rawPage, ok := r.Form["page"]
		if ok {
			req.Page, err = strconv.Atoi(rawPage[0])
			if err != nil {
				log.Error("failed to make int page", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to make int page"))
				return
			}
		} else {
			req.Page = 1
		}

		// Запрашиваем все посты из БД
		var p news.Post
		ps, err := p.GetPostsPage(storage, req.Tags, req.Page)
		// Случай когда указана страница вне диапазона
		if errors.As(err, &(storageHandler.ErrPageInOutOfRange)) {
			log.Debug("failed to get catalog", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("selected page in out of range"))
			return
		}
		// Общий случай ошибки
		if err != nil {
			log.Error("failed to get posts", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get posts"))
			return
		}

		// Записываем все посты из БД в структуру ответа на запрос
		var articles []Article
		for _, post := range ps {
			if len(post.Text) > 64 {
				post.Text = post.Text[:64] // Вырезка первых 64 символов новости для превью
			}
			a := Article{Post: post}
			articles = append(articles, a)
		}

		// Записываем оставшиеся поля структуры ответа
		for i := range articles {
			// Запрос и запись количества лайков для поста
			var l news.Like
			amount, err := l.GetLikesAmount(storage, articles[i].PostID)
			if err != nil {
				log.Error("failed to get likes amount", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get likes amount"))
				return
			}
			articles[i].LikesAmount = amount

			// Запрос и запись путей к изображениям для поста
			var pi news.PostImage
			paths, err := pi.GetImagePathsByPostID(storage, articles[i].PostID)
			if err != nil {
				log.Error("failed to get post image paths", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get post image paths"))
				return
			}
			articles[i].Images = paths

			// Запрос и запись тэгов для поста
			var t news.Tag
			tags, err := t.GetTagsByPostID(storage, articles[i].PostID)
			if err != nil {
				log.Error("failed to get post tags", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get post tags"))
				return
			}
			articles[i].Tags = tags
		}

		log.Info("articles successfully gotten")

		responseOK(w, r, log, articles)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, articles []Article) {
	response, err := json.Marshal(Response{
		Response: resp.OK(),
		Articles: articles,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
