package drop_cart

import (
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.addCartItem.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]string)
		userID, err := strconv.Atoi(tempUserID["user_id"])
		if err != nil {
			log.Error("failed to get user id from token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("failed to get user id from token claims"))
			return
		}

		// Запрос и проверка доступности item для заказа
		var c *shop.Cart
		err = c.DropCart(storage, userID)

		// Обработка общего случая ошибки БД
		if err != nil {
			log.Error(err.Error())

			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to drop item"))

			return
		}

		log.Info("item successfully added in cart")

		render.JSON(w, r, resp.OK())
	}
}
