package cartData

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	storageHandler "portal/internal/storage"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
)

type Response struct {
	resp.Response
	InCartItems []shop.InCartItem `json:"cart_data"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.getCartData.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем userID из токена авторизации
		tempUserID := r.Context().Value(oauth.ClaimsContext).(map[string]int)
		userID, ok := tempUserID["user_id"]
		if !ok {
			log.Error("no user id in token claims")
			w.WriteHeader(500)
			render.JSON(w, r, resp.Error("no user id in token claims"))
			return
		}

		// Запрос cart_id для вызывающего user_id
		var c shop.Cart
		err := c.GetActiveCartID(storage, userID)
		if err != nil {
			// Если ошибка не об отсутствии корзины, то выход по стнадартной ошибке БД
			if !errors.As(err, &storageHandler.ErrCartDoesNotExist) {
				log.Error("failed to get active cart id", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get active cart id"))
				return
			}
			// Если ошибка выше была об отсутствии корзины, то создаем корзину
			if err := c.NewCart(storage, userID); err != nil {
				log.Error("failed to create new cart", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to create cart"))
				return
			}
			// Получаем номер созданной корзины
			if err := c.GetActiveCartID(storage, userID); err != nil {
				log.Error("failed to get active cart id", sl.Err(err))
				w.WriteHeader(422)
				render.JSON(w, r, resp.Error("failed to get active cart id"))
				return
			}
		}

		// Заполняем слайс товарами для нужной корзины из БД
		var ici *shop.InCartItem
		icis, err := ici.GetInCartItems(storage, c.CartID)
		if err != nil {
			log.Error("failed to get in cart items", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get in cart items"))
			return
		}

		log.Info("cart data loaded")

		responseOK(w, r, log, icis)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, inCartItems []shop.InCartItem) {
	response, err := json.Marshal(Response{
		Response:    resp.OK(),
		InCartItems: inCartItems,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
