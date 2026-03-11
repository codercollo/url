package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url/internal/config"
	"url/internal/handler"
	"url/internal/mailer"
	"url/internal/repository"
	"url/internal/routes"
	"url/internal/service"
	"url/pkg/logger"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	// Set Gin mode from config
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Postgres
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		log.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}
	log.Info("postgres connected!")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Error("redis ping failed", "error", err)
		os.Exit(1)
	}
	log.Info("redis connected")

	// Session manager
	sessionManager := scs.New()
	sessionManager.Lifetime = cfg.Session.Lifetime
	sessionManager.Cookie.HttpOnly = cfg.Session.HttpOnly
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = cfg.Session.SecureCookie

	//Mailer
	m, err := mailer.New(mailer.Config{
		Host:     cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		Username: cfg.Mail.Username,
		Password: cfg.Mail.Password,
		From:     cfg.Mail.From,
	})
	if err != nil {
		log.Error("failed to initialize mailer", "error", err)
		os.Exit(1)
	}
	// Worker pool — starts cfg.Mail.Workers goroutines immediately
	mailWorker := mailer.NewWorkerPool(m, cfg.Mail.Workers, cfg.Mail.QueueSize)
	defer mailWorker.Shutdown()

	// Repository layer
	pgStore := repository.NewPostgresStore(db)
	redisCache := repository.NewRedisCache(rdb, cfg.Redis.CacheTTL)

	// Service layer
	shortenerSvc := service.NewShortenerService(pgStore, redisCache, log)
	analyticsSvc := service.NewAnalyticsService(pgStore)
	authSvc := service.NewAuthSevice(pgStore, mailWorker, cfg.AppBaseURL)

	// Handlers
	h := handler.NewHandler(
		shortenerSvc,
		analyticsSvc,
		authSvc,
		sessionManager,
		db,
		rdb,
		mailWorker,
		cfg.AppBaseURL,
		log,
	)

	// Server
	r := gin.Default()
	r.SetTrustedProxies([]string{cfg.Server.TrustedProxy})
	routes.Register(r, h, pgStore, sessionManager)

	//HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	//Start server in a goroutine
	go func() {
		log.Info("server starting", "addr", cfg.Server.Port, "emv", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error!", "error", err)
			os.Exit(1)
		}
	}()

	//Block until SIGINT or SIGTERM received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	if err := rdb.Close(); err != nil {
		log.Error("redis close error", "error", err)
	}

	log.Info("server stopped cleanly")

}
