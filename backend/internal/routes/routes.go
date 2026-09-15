package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avanthika/efootball-backend/internal/handlers"
	"github.com/avanthika/efootball-backend/internal/repository"
	"github.com/avanthika/efootball-backend/internal/service"
)

func Register(r *gin.Engine, db *pgxpool.Pool) {
	healthRepo := repository.NewHealthRepository(db)
	healthService := service.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthService)

	r.GET("/health", healthHandler.Check)
}
