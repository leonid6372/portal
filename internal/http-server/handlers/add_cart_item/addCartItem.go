package addCartItem

import (
	"errors"
	"io"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/shop"
	"strconv"

	"log/slog"

	storageHandler "portal/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	ItemID   int `json:"item_id,omitempty" validate:"required"`
	Quantity int `json:"quantity,omitempty" validate:"required"`
}

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
			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

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
		var i shop.Item
		if err := i.GetIsAvailable(storage, req.ItemID); err != nil {
			log.Error("failed to get item status", err)
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get item status: "+err.Error()))
			return
		}
		if !i.IsAvailable {
			log.Error("item is not available")
			w.WriteHeader(406)
			render.JSON(w, r, resp.Error("item is not available"))
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

		// Добавление item в корзину
		var ici shop.InCartItem
		if err := ici.NewInCartItem(storage, req.ItemID, req.Quantity, c.CartID); err != nil {
			log.Error("failed to add item in cart", err)
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to add item in cart: "+err.Error()))
			return
		}

		log.Info("item successfully added in cart")

		render.JSON(w, r, resp.OK())
	}
}
