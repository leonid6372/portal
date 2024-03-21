package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"portal/internal/config"

	addCartItem "portal/internal/http-server/handlers/add_cart_item"
	getReservationList "portal/internal/http-server/handlers/get_reservation_list"
	getShopList "portal/internal/http-server/handlers/get_shop_list"
	reservationHandler "portal/internal/http-server/handlers/reservation"

	"portal/internal/lib/auth"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	logLVLInfo  = "info"
	logLVLDebug = "debug"
	logLVLWarn  = "warning"
	logLVLError = "error"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.LogLVL)

	storage, err := postgres.New(cfg.SQLStorage)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	err = auth.InitBearerServer(log, storage, cfg.TokenTTL)
	if err != nil {
		log.Error("failed to init bearer server", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Logger)    // Логирование всех запросов
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов

	routeAPI(router, log, storage)

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

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(logLVL string) *slog.Logger {
	var log *slog.Logger

	switch logLVL {
	case logLVLInfo:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case logLVLDebug:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case logLVLWarn:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case logLVLError:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	return log
}

func routeAPI(router *chi.Mux, log *slog.Logger, storage *postgres.Storage) {
	//Secured API group
	router.Group(func(r chi.Router) {
		// use the Bearer Authentication middleware
		r.Use(auth.GetAuthHandler(log))
		r.Post("/add_cart_item", addCartItem.New(log, storage)) // TO DO: переделать под новые поля таблицы in_cart_item
		r.Get("/get_shop_list", getShopList.New(log, storage))
		r.Post("/reservation", reservationHandler.New(log, storage))
		r.Get("/get_reservation_list", getReservationList.New(log, storage))
	})

	// Public API group
	router.Group(func(r chi.Router) {
		r.Post("/login", auth.GetBearerServer().UserCredentialsPassword)
		r.Post("/refresh", auth.GetBearerServer().UserCredentialsRefresh)
	})
}
