package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avanthika/efootball-backend/internal/handlers"
	"github.com/avanthika/efootball-backend/internal/middleware"
	"github.com/avanthika/efootball-backend/internal/repository"
	"github.com/avanthika/efootball-backend/internal/service"
)

func Register(r *gin.Engine, db *pgxpool.Pool, jwtSecret string) {
	healthRepo := repository.NewHealthRepository(db)
	healthService := service.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthService)
	r.GET("/health", healthHandler.Check)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handlers.NewAuthHandler(authService)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/signup", authHandler.Signup)

	protected := r.Group("/")
	protected.Use(middleware.AuthRequired(jwtSecret))
	protected.GET("/me", handlers.Me)

	tournamentRepo := repository.NewTournamentRepository(db)
	spinWheelRepo := repository.NewSpinWheelRepository(db)
	spinWheelService := service.NewSpinWheelService(tournamentRepo, spinWheelRepo)
	spinWheelHandler := handlers.NewSpinWheelHandler(spinWheelService)

	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService, spinWheelService)

	adminOnly := r.Group("/admin")
	adminOnly.Use(middleware.AuthRequired(jwtSecret), middleware.RequireRole("admin"))
	adminOnly.GET("/subadmins", userHandler.ListSubadmins)
	adminOnly.GET("/subadmins/:id", userHandler.GetSubadmin)
	adminOnly.DELETE("/subadmins/:id", userHandler.DeleteSubadmin)

	// Public — no login required, reached via the shareable registration link.
	r.GET("/register/:token", spinWheelHandler.GetPublic)
	r.POST("/register/:token", spinWheelHandler.Register)

	tournamentOwner := r.Group("/tournaments")
	tournamentOwner.Use(middleware.AuthRequired(jwtSecret), middleware.RequireRole("subadmin", "admin"))
	tournamentOwner.POST("", spinWheelHandler.Create)
	tournamentOwner.GET("", spinWheelHandler.ListMine)
	tournamentOwner.GET("/:id", spinWheelHandler.Get)
	tournamentOwner.PATCH("/:id", spinWheelHandler.UpdateBracketSize)
	tournamentOwner.DELETE("/:id", spinWheelHandler.Delete)
	tournamentOwner.POST("/:id/entries", spinWheelHandler.AddManualEntry)
	tournamentOwner.GET("/:id/entries", spinWheelHandler.ListEntries)
	tournamentOwner.PATCH("/:id/entries/:entryId", spinWheelHandler.UpdateEntry)
	tournamentOwner.DELETE("/:id/entries/:entryId", spinWheelHandler.DeleteEntry)
	tournamentOwner.POST("/:id/assign-colors", spinWheelHandler.AssignColors)
	tournamentOwner.POST("/:id/spin", spinWheelHandler.StartSpin)
	tournamentOwner.GET("/:id/fixtures", spinWheelHandler.GetFixtures)
	tournamentOwner.PATCH("/:id/matches/:matchId", spinWheelHandler.SubmitResult)
}
