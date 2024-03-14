package main

import (
	"context"
	"github.com/go-chi/jwtauth/v5"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"portal/internal/config"
	addCartItem "portal/internal/http-server/handlers/add_cart_item"
	"portal/internal/http-server/handlers/get_reservation_list"
	getShopList "portal/internal/http-server/handlers/get_shop_list"
	logIn "portal/internal/http-server/handlers/log_in"
	"portal/internal/http-server/handlers/reservation"
	"portal/internal/lib/jwt"
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

var (
	tokenAuth *jwtauth.JWTAuth
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.LogLVL)

	storage, err := postgres.New(cfg.SQLStorage)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Logger)    // Логирование всех запросов
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов

	// Инициализация шаблона для построения токенов по секрету и типу шифрования
	tokenAuth, _ = jwt.Init()

	// Protected routes
	router.Group(func(router chi.Router) {
		// Seek, verify and validate JWT tokens
		router.Use(jwtauth.Verifier(tokenAuth))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		router.Use(jwtauth.Authenticator(tokenAuth))

		/* EXAMPLE: router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
		})*/

	})

	// Public routes
	router.Group(func(r chi.Router) {
		/* EXAMPLE: router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})*/
		router.Get("/api/log_in", logIn.New(log, storage, tokenAuth))
		router.Post("/api/add_cart_item", addCartItem.New(log, storage))
		router.Get("/api/get_shop_list", getShopList.New(log, storage))
		router.Post("/api/reservation", reservation.New(log, storage))
		router.Get("/api/get_reservation_list", getReservationList.New(log, storage))
	})

	/*router.Get("/api/get_shop_list", getShopList.New(log, storage))
	router.Post("/api/add_cart_item", addCartItem.New(log, storage))*/

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
