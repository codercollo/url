package main

import (
	"context"
	"database/sql"
	"log"
	"time"
	"url/internal/handler"
	"url/internal/repository"
	"url/internal/routes"
	"url/internal/service"

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

	//Layers
	pgStore := repository.NewPostgresStore(db)
	redisCache := repository.NewRedisCache(rdb, 24*time.Hour)
	shortenerSvc := service.NewShortenerService(pgStore, redisCache)

	//Handlers
	h := handler.NewHandler(shortenerSvc)

	//Server
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	//pgStore implements both URLStore and ClickStore
	routes.Register(r, h, pgStore)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to start server: %v", err)
	}
}
