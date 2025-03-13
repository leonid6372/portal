package articles

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	vc "portal/internal/lib/views_counter"
	storageHandler "portal/internal/storage"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"strconv"
	"time"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const (
	previewWordsAmount = 10
)

var (
	errPageInOutOfRange = errors.New("page in out of range")
)

type Request struct {
	TagsID        []string //string потому что потом будетв вставляться в sql запрос и там нужен string
	Page          int
	CreatedAfter  time.Time
	CreatedBefore time.Time
	UserID        int
}

// Запрашиваемая API структура
type Article struct {
	news.Post
	LikesAmount    int        `json:"likes_amount"`
	CommentsAmount int        `json:"comments_amount"`
	IsLiked        bool       `json:"is_liked"`
	Images         []string   `json:"images"`
	Tags           []news.Tag `json:"tags"`
}

type Response struct {
	resp.Response
	Articles   []Article `json:"articles"`
	TotalViews int       `json:"total_views"`
}

func New(log *slog.Logger, storage *postgres.Storage, viewsCounter *vc.ViewsCounter) http.HandlerFunc {
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
		req.TagsID = r.Form["tag_id"]
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
		rawCreatedAfter, ok := r.Form["start"]
		if ok {
			intCreatedAfter, err := strconv.Atoi(rawCreatedAfter[0][:len(rawCreatedAfter[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert reservation start time", sl.Err(err))
			}
			req.CreatedAfter = time.Unix(int64(intCreatedAfter), 0)
		} else { // Если не задана начальная дата создания поста для фильтра, то выводим все от самого начала
			req.CreatedAfter = time.Time{}
		}
		rawCreatedBefore, ok := r.Form["finish"]
		if ok {
			intCreatedBefore, err := strconv.Atoi(rawCreatedBefore[0][:len(rawCreatedBefore[0])-3]) // Cut three time zone zeroes at the end
			if err != nil {
				log.Error("failed to convert reservation finish time", sl.Err(err))
			}
			req.CreatedBefore = time.Unix(int64(intCreatedBefore), 0)
		} else { // Если не задана крайняя дата создания поста для фильтра, то выводим все до сейчас
			req.CreatedBefore = time.Now()
		}
		rawUserID, ok := r.Form["user_id"]
		if ok {
			req.UserID, err = strconv.Atoi(rawUserID[0])
			if err != nil {
				log.Error("failed to make int user_id", sl.Err(err))
				w.WriteHeader(500)
				render.JSON(w, r, resp.Error("failed to make int user_id"))
				return
			}
		}

		// Запрашиваем все посты из БД
		var p news.Post
		ps, err := p.GetPostsPage(storage, req.TagsID, req.Page, req.CreatedAfter, req.CreatedBefore)
		// Случай когда указана страница вне диапазона
		if errors.As(err, &(storageHandler.ErrPageInOutOfRange)) {
			log.Error("failed to get catalog", sl.Err(err))
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
			compressedText := ""
			spaceAmount := 0
			for _, rn := range post.Text { // Вырезка первых 10 слов
				compressedText += string(rn)
				if rn == ' ' {
					spaceAmount++
				}
				if spaceAmount == 10 {
					break
				}
			}
			post.Text = compressedText + "..."
			a := Article{Post: post}
			articles = append(articles, a)
		}

		// Записываем оставшиеся поля структуры ответа
		for i := range articles {
			// Запрос и запись количества лайков для поста
			var l news.Like
			likesAmount, err := l.GetLikesAmount(storage, articles[i].PostID)
			if err != nil {
				log.Error("failed to get likes amount", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get likes amount"))
				return
			}
			articles[i].LikesAmount = likesAmount

			isLiked := true
			if req.UserID != 0 {
				isLiked, err = l.IsLikedByUserID(storage, articles[i].PostID, req.UserID)
				if err != nil {
					log.Error("failed to check post is liked by user", sl.Err(err))
					w.WriteHeader(422)
					render.JSON(w, r, resp.Error("failed to check post is liked by user"))
					return
				}
			}
			articles[i].IsLiked = isLiked

			// Получаем кол-во комментариев по post ID
			var c news.Comment
			commentsAmount, err := c.GetCommentsAmount(storage, articles[i].PostID)
			if err != nil {
				log.Error("failed to get comments amount", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get comments amount"))
				return
			}
			articles[i].CommentsAmount = commentsAmount

			// Запрос и запись путей к изображениям для поста
			var pi news.PostImage
			paths, err := pi.GetImageInfoByPostID(storage, articles[i].PostID)
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

		// Читаем значение просмотров всего было при запуске сервера из БД
		views, err := storage.GetViews()
		if err != nil {
			log.Error("failed to get views from postgres", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get views from postgres"))
			return
		}

		// Добавляем к значению просмотров всего было просмотры во время текущего сеасна работы сервера
		curSessionViews := viewsCounter.Count()
		totalViews := views + curSessionViews

		responseOK(w, r, log, articles, totalViews)

		viewsCounter.Add(1)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, articles []Article, totalViews int) {
	response, err := json.Marshal(Response{
		Response:   resp.OK(),
		Articles:   articles,
		TotalViews: totalViews,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
