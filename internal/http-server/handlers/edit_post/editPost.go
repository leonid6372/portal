package editPost

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	minioServer "portal/internal/storage/minio"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/news"
	"portal/internal/structs/models"
	"portal/internal/structs/roles"
	"slices"

	resp "portal/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	PostID int    `json:"post_id" validate:"required"`
	Title  string `json:"title" validate:"required"`
	Text   string `json:"text" validate:"required"`
	Tags   []int  `json:"tags" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage, miniosrv *minioServer.MinioProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.editPost.New"

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

		// Декодируем данные из запроса в json
		data := r.FormValue("post_data")
		if data == "" {
			log.Error("request body is empty")
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("empty request"))
			return
		}
		err := json.Unmarshal([]byte(data), &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		allowedImageExtensions := []string{".png", ".jpg", ".jpeg"}
		maxImageSize := int64(9437184) // 9 MB

		// Забираем фото из тела запроса. Если нет фото, то нет ошибки.
		src, hdr, err := r.FormFile("image")
		if err != nil && err.Error() != "request Content-Type isn't multipart/form-data" {
			log.Error("failed to get image from request body", sl.Err(err))
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("failed to get image from request body"))
			return
		}

		var imageExtension, imageName string
		var imageNumber int

		// Если фото есть, проверям его соответсвие требованиям.
		if src != nil {
			defer src.Close()

			imageExtension = filepath.Ext(hdr.Filename)
			if !slices.Contains(allowedImageExtensions, imageExtension) {
				log.Error("image extension is not allowed")
				w.WriteHeader(400)
				render.JSON(w, r, resp.Error("image extension is not allowed"))
				return
			}

			if hdr.Size > maxImageSize {
				log.Error("image size out of limit")
				w.WriteHeader(400)
				render.JSON(w, r, resp.Error("image size out of limit"))
				return
			}
		}

		// Удаляем изображения поста из БД. Будет возвращать число фото у поста
		var pi news.PostImage
		if err := pi.DeletePostImageByPostID(storage, req.PostID); err != nil {
			log.Error("failed to delete post images", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to delete post images"))
			return
		}

		// Удаляем фото из MinIO
		imageNumber = 1 // TO DO: Переделать для много фото, чтобы считалось
		imageName = fmt.Sprintf("post_images/post%d_image%d", req.PostID, imageNumber)

		// Удаляем фото из хранилища
		err = miniosrv.RemoveImage(r.Context(), imageName)
		if err != nil {
			log.Error("failed to remove image from minio", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to remove image from minio"))
			return
		}

		// Удаляем тэги к посту
		var ipt news.InPostTag
		if err := ipt.DeleteInPostTagByPostID(storage, req.PostID); err != nil {
			log.Error("failed to delete post tags", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to delete post tags"))
			return
		}

		// Обновляем пост в БД
		var p news.Post
		if err := p.UpdatePost(storage, req.Title, req.Text, req.PostID); err != nil {
			log.Error("failed to update post", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to update post"))
			return
		}

		// Добавляем информацию о новых изображениях к посту
		if err := pi.NewPostImage(storage, req.PostID, imageName); err != nil {
			log.Error("failed to add image to post", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to add image to post"))
			return
		}

		// Загружаем фото в MinIO
		imageNumber = 1 // TO DO: Переделать для много фото, чтобы считалось
		if src != nil {
			image := models.Image{
				Payload:   src,
				Name:      imageName,
				Size:      hdr.Size,
				Extension: imageExtension,
			}

			// Отправляем фото в хранилище
			err = miniosrv.UploadImage(r.Context(), image)
			if err != nil {
				log.Error("failed to upload image to minio", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to upload image to minio"))
				return
			}
		}

		// Добавляем новые тэги к посту
		for _, tag := range req.Tags {
			if err := ipt.NewInPostTag(storage, req.PostID, tag); err != nil {
				log.Error("failed to add tag to post", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to add tag to post"))
				return
			}
		}

		log.Info("article successfully edited")

		render.JSON(w, r, resp.OK())
	}
}
