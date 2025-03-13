package image

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	minioServer "portal/internal/storage/minio"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Name string
}

type Response struct {
	resp.Response
}

func New(log *slog.Logger, miniosrv *minioServer.MinioProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.image.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Считываем параметры запроса из request
		r.ParseForm()
		name, ok := r.Form["name"]
		if !ok {
			log.Error("empty image name parameter")
			w.WriteHeader(400)
			render.JSON(w, r, resp.Error("empty image name parameter"))
			return
		}
		req.Name = name[0]

		file, err := miniosrv.DownloadImage(r.Context(), req.Name)
		if err != nil && err.Error() == "Cant send photo: The specified key does not exist." {
			log.Error("image with this name doesn't exist")
			w.WriteHeader(406)
			render.JSON(w, r, resp.Error("image with this name doesn't exist"))
			return
		}
		if err != nil {
			log.Error("failed to get image from minio")
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get image from minio"))
			return
		}
		defer file.Close()

		log.Info("image successfully gotten")

		w.Header().Set("Content-Type", "image/png")
		if _, err := io.Copy(w, file); err != nil {
			fmt.Printf("Cant send photo: %v\n", err)
			http.Error(w, "Can`t download photo!", http.StatusInternalServerError)
			return
		}

		log.Info("image successfully set to response")

		render.JSON(w, r, resp.OK())
	}
}
