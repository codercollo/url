package main

import (
	"context"
	"database/sql"
	"log"
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

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// Set Gin mode from config
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Postgres
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		log.Fatal("postgres ping failed:", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping failed:", err)
	}

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
		log.Fatal("failed to init mailer:", err)
	}
	// Worker pool — starts cfg.Mail.Workers goroutines immediately
	mailWorker := mailer.NewWorkerPool(m, cfg.Mail.Workers, cfg.Mail.QueueSize)
	defer mailWorker.Shutdown()

	// Repository layer
	pgStore := repository.NewPostgresStore(db)
	redisCache := repository.NewRedisCache(rdb, cfg.Redis.CacheTTL)

	// Service layer
	shortenerSvc := service.NewShortenerService(pgStore, redisCache)
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
		log.Printf("Starting server on %s (env: %s)", cfg.Server.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	//Block until SIGINT or SIGTERM received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("server stopped cleanly")

}
