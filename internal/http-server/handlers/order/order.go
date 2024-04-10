package order

import (
	"errors"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/oauth"
	storageHandler "portal/internal/storage"
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
		const op = "handlers.order.New"

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
			render.JSON(w, r, resp.Error("failed to get user id from token claim: "+err.Error()))
			return
		}

		// Запрос cart_id для вызывающего user_id
		var c shop.Cart
		err = c.GetActiveCartID(storage, userID)
		if err != nil {
			// Если ошибка не об отсутствии корзины, то выход по стнадартной ошибке БД
			if !errors.As(err, &storageHandler.ErrCartDoesNotExist) {
				log.Error("failed to get active cart id", err)
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get active cart id: "+err.Error()))
				return
			}
			// Если ошибка выше была об отсутствии корзины, то создаем корзину
			if err := c.NewCart(storage, userID); err != nil {
				log.Error("failed to create new cart", err)
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to create cart: "+err.Error()))
				return
			}
			// Получаем номер созданной корзины
			if err := c.GetActiveCartID(storage, userID); err != nil {
				log.Error("failed to get active cart id", err)
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get active cart id: "+err.Error()))
				return
			}
		}

		// Переводим корзину с cartID в неактивное состояние
		if err := c.UpdateCartToInactive(storage, c.CartID); err != nil {
			log.Error("failed to make order", err)
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to make order: "+err.Error()))
			return
		}

		log.Info("order successfully made")

		render.JSON(w, r, resp.OK())
	}
}
