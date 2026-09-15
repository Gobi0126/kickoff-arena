package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/avanthika/efootball-backend/internal/config"
	"github.com/avanthika/efootball-backend/internal/db"
	"github.com/avanthika/efootball-backend/internal/routes"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pgPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}
	log.Println("connected to postgres")
	defer pgPool.Close()

	redisClient, err := db.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}
	log.Println("connected to redis")
	defer redisClient.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	routes.Register(r, pgPool)

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
