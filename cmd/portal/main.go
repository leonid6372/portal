package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"portal/internal/config"
	addCartItem "portal/internal/http-server/handlers/add_cart_item"
	approveComment "portal/internal/http-server/handlers/approve_comment"
	"portal/internal/http-server/handlers/article"
	"portal/internal/http-server/handlers/articles"
	cartData "portal/internal/http-server/handlers/cart_data"
	checkComments "portal/internal/http-server/handlers/check_comments"
	"portal/internal/http-server/handlers/comment"
	createPost "portal/internal/http-server/handlers/create_post"
	deleteComment "portal/internal/http-server/handlers/delete_comment"
	deleteItem "portal/internal/http-server/handlers/delete_item"
	deletePost "portal/internal/http-server/handlers/delete_post"
	dropCart "portal/internal/http-server/handlers/drop_cart"
	dropCartItem "portal/internal/http-server/handlers/drop_cart_item"
	editComment "portal/internal/http-server/handlers/edit_comment"
	editPost "portal/internal/http-server/handlers/edit_post"
	"portal/internal/http-server/handlers/like"
	lockerReservation "portal/internal/http-server/handlers/locker_reservation"
	lockerReservationDrop "portal/internal/http-server/handlers/locker_reservation_drop"
	lockerReservationList "portal/internal/http-server/handlers/locker_reservation_list"
	lockerReservationUpdate "portal/internal/http-server/handlers/locker_reservation_update"
	"portal/internal/http-server/handlers/me"
	"portal/internal/http-server/handlers/order"
	profile "portal/internal/http-server/handlers/profile"
	reservationHandler "portal/internal/http-server/handlers/reservation"
	reservationDelete "portal/internal/http-server/handlers/reservation_delete"
	reservationDrop "portal/internal/http-server/handlers/reservation_drop"
	reservationEdit "portal/internal/http-server/handlers/reservation_edit"
	reservationList "portal/internal/http-server/handlers/reservation_list"
	reservationUpdate "portal/internal/http-server/handlers/reservation_update"
	shopList "portal/internal/http-server/handlers/shop_list"
	"portal/internal/http-server/handlers/tag"
	tags "portal/internal/http-server/handlers/tags"
	updateCartItem "portal/internal/http-server/handlers/update_cart_item"
	userLockerReservations "portal/internal/http-server/handlers/user_locker_reservations"
	userReservations "portal/internal/http-server/handlers/user_reservations"

	setupLogger "portal/internal/lib/logger/setup_logger"
	"portal/internal/lib/logger/sl"
	"portal/internal/lib/oauth"
	ldapServer "portal/internal/storage/ldap"
	"portal/internal/storage/mssql"
	"portal/internal/storage/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log, logFile := setupLogger.New(cfg.LogLVL)
	defer logFile.Close()

	// Подключение и закрытие базы PSQL
	storage, err := postgres.New(cfg.SQL)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		storage.DB.Close()
		log.Info("storage closed")
	}()

	// Подключение и закрытие базы MSSQL 1C
	storage1C, err := mssql.New(cfg.SQL)
	if err != nil {
		log.Error("failed to init 1C storage", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		storage1C.DB.Close()
		log.Info("1C storage closed")
	}()

	// Подключение и закрытие LDAP сервера
	ldapsrv, err := ldapServer.New(cfg.LDAPServer.FQDN, cfg.LDAPServer.BaseDN, cfg.LDAPServer.UserAccountControl)
	if err != nil {
		log.Error("failed to init LDAP server", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		ldapsrv.LDAPConn.Close()
		log.Info("LDAP server closed")
	}()

	bearerServer := oauth.NewBearerServer(
		cfg.BearerServer.Secret,
		cfg.BearerServer.TokenTTL,
		&oauth.UserVerifier{Storage: storage /*Storage1C: storage1C*/, LDAPServer: ldapsrv, Log: log},
		nil)

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Добавляет request_id в каждый запрос, для трейсинга
	router.Use(middleware.Logger)    // Логирование всех запросов
	router.Use(middleware.Recoverer) // Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов

	routeAPI(router, log, bearerServer, cfg.BearerServer.Secret, storage, storage1C)

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

func routeAPI(router *chi.Mux, log *slog.Logger, bearerServer *oauth.BearerServer, secret string, storage *postgres.Storage, storage1C *mssql.Storage) {
	//Secured API group
	router.Group(func(r chi.Router) {
		// use the Bearer Authentication middleware
		r.Use(oauth.Authorize(secret, nil, bearerServer, log))
		r.Post("/api/reservation", reservationHandler.New(log, storage))
		r.Get("/api/reservation_list", reservationList.New(log, storage, storage1C))
		r.Get("/api/user_reservations", userReservations.New(log, storage))
		r.Post("/api/reservation_update", reservationUpdate.New(log, storage))
		r.Post("/api/reservation_drop", reservationDrop.New(log, storage))

		r.Post("/api/reservation_delete", reservationDelete.New(log, storage))
		r.Post("/api/reservation_edit", reservationEdit.New(log, storage))

		r.Post("/api/locker_reservation", lockerReservation.New(log, storage))
		r.Get("/api/locker_reservation_list", lockerReservationList.New(log, storage, storage1C))
		r.Get("/api/user_locker_reservations", userLockerReservations.New(log, storage))
		r.Post("/api/locker_reservation_update", lockerReservationUpdate.New(log, storage))
		r.Post("/api/locker_reservation_drop", lockerReservationDrop.New(log, storage))

		r.Get("/api/profile", profile.New(log, storage, storage1C))
		r.Get("/api/me", me.New(log, storage))

		r.Get("/api/shop_list", shopList.New(log, storage))
		r.Post("/api/add_cart_item", addCartItem.New(log, storage))
		r.Post("/api/order", order.New(log, storage))
		r.Get("/api/cart_data", cartData.New(log, storage))
		r.Post("/api/drop_cart", dropCart.New(log, storage))
		r.Post("/api/drop_cart_item", dropCartItem.New(log, storage))
		r.Post("/api/update_cart_item", updateCartItem.New(log, storage))
		r.Post("/api/delete_item", deleteItem.New(log, storage))

		r.Post("/api/comment", comment.New(log, storage))
		r.Post("/api/edit_comment", editComment.New(log, storage))
		r.Get("/api/check_comments", checkComments.New(log, storage, storage1C))
		r.Post("/api/approve_comment", approveComment.New(log, storage))
		r.Post("/api/delete_comment", deleteComment.New(log, storage))

		r.Post("/api/like", like.New(log, storage))

		r.Post("/api/create_article", createPost.New(log, storage))
		r.Post("/api/edit_article", editPost.New(log, storage))
		r.Post("/api/delete_article", deletePost.New(log, storage))

		r.Post("/api/tag", tag.New(log, storage))
		r.Post("/api/delete_tag", tag.New(log, storage))
		r.Get("/api/tags", tags.New(log, storage))
	})

	// Public API group
	router.Group(func(r chi.Router) {
		r.Post("/api/login", bearerServer.UserCredentials)

		r.Get("/api/articles", articles.New(log, storage))
		r.Get("/api/article", article.New(log, storage, storage1C))
	})
}
