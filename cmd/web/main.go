package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"
	"url/internal/handler"
	"url/internal/repository"
	"url/internal/routes"
	"url/internal/service"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	//Postgres
	db, err := sql.Open("postgres", "postgres://urluser:urlpassword@localhost:5432/urldb?sslmode=disable")
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}
	defer db.Close()

	//Verify Postgres is alive
	if err := db.Ping(); err != nil {
		log.Fatal("postgres ping failed:", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//Verify Redis is alive
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping failed:", err)
	}

	//Session Manager {scs}
	//Uses a plain cookie store by default
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	//Repository Layer
	pgStore := repository.NewPostgresStore(db)
	redisCache := repository.NewRedisCache(rdb, 24*time.Hour)

	//Service Layer
	shortenerSvc := service.NewShortenerService(pgStore, redisCache)
	analyticsSvc := service.NewAnalyticsService(pgStore)
	authSvc := service.NewAuthSevice(pgStore)
	//Handlers
	h := handler.NewHandler(shortenerSvc, analyticsSvc, authSvc, sessionManager)

	//Server
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	//pgStore implements both URLStore and ClickStore
	routes.Register(r, h, pgStore, sessionManager)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
