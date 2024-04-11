package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"portal/internal/config"
	addCartItem "portal/internal/http-server/handlers/add_cart_item"
	dropCart "portal/internal/http-server/handlers/drop_cart"
	dropCartItem "portal/internal/http-server/handlers/drop_cart_item"
	cartData "portal/internal/http-server/handlers/get_cart_data"
	"portal/internal/http-server/handlers/order"
	profile "portal/internal/http-server/handlers/profile"
	reservationHandler "portal/internal/http-server/handlers/reservation"
	reservationDrop "portal/internal/http-server/handlers/reservation_drop"
	reservationList "portal/internal/http-server/handlers/reservation_list"
	reservationUpdate "portal/internal/http-server/handlers/reservation_update"
	shopList "portal/internal/http-server/handlers/shop_list"
	updateCartItem "portal/internal/http-server/handlers/update_cart_item"
	userReservations "portal/internal/http-server/handlers/user_reservations"

	setupLogger "portal/internal/lib/logger/setup_logger"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	"portal/internal/storage/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger.New(cfg.LogLVL)

	storage, err := postgres.New(cfg.SQLStorage)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		storage.DB.Close()
		log.Info("storage closed")
	}()

	bearerServer := oauth.NewBearerServer(
		cfg.BearerServer.Secret,
		cfg.BearerServer.TokenTTL,
		&oauth.UserVerifier{Storage: storage, Log: log},
		nil)

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Logger)    // Логирование всех запросов
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов

	routeAPI(router, log, bearerServer, cfg.BearerServer.Secret, storage)

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server")

		return
	}

	log.Info("server stopped")
}

func routeAPI(router *chi.Mux, log *slog.Logger, bearerServer *oauth.BearerServer, secret string, storage *postgres.Storage) {
	//Secured API group
	router.Group(func(r chi.Router) {
		// use the Bearer Authentication middleware
		r.Use(oauth.Authorize(secret, nil, bearerServer, log))
		r.Post("/api/reservation", reservationHandler.New(log, storage))
		r.Get("/api/reservation_list", reservationList.New(log, storage))
		r.Get("/api/user_reservations", userReservations.New(log, storage))
		r.Post("/api/reservation_update", reservationUpdate.New(log, storage))
		r.Post("/api/reservation_drop", reservationDrop.New(log, storage))

		r.Get("/api/profile", profile.New(log, storage))

		r.Get("/api/shop_list", shopList.New(log, storage))
		r.Post("/api/add_cart_item", addCartItem.New(log, storage))
		r.Post("/api/order", order.New(log, storage))
		r.Get("/api/cart_data", cartData.New(log, storage))
		r.Post("/api/drop_cart", dropCart.New(log, storage))
		r.Post("/api/drop_cart_item", dropCartItem.New(log, storage))
		r.Post("/api/update_cart_item", updateCartItem.New(log, storage))

	})

	// Public API group
	router.Group(func(r chi.Router) {
		r.Post("/api/login", bearerServer.UserCredentials)
	})
}
