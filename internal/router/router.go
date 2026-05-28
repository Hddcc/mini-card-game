package router

import (
	"mini-card-game/internal/config"
	"mini-card-game/internal/handler"
	"mini-card-game/internal/middleware"
	"mini-card-game/internal/repository"
	"mini-card-game/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Logger *zap.Logger
}

func New(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repository.NewUserRepository(deps.DB)
	playerRepo := repository.NewPlayerRepository(deps.DB)
	authService := service.NewAuthService(deps.Config, deps.DB, userRepo, playerRepo)
	authHandler := handler.NewAuthHandler(authService)
	playerService := service.NewPlayerService(playerRepo)
	playerHandler := handler.NewPlayerHandler(playerService)

	api := r.Group("/api/v1")
	loginRequired := api.Group("")
	loginRequired.Use(middleware.Auth(deps.Config.JWTSecret))
	loginRequired.GET("/player/profile", playerHandler.Profile)
	loginRequired.GET("/player/assets", playerHandler.Assets)

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	return r
}
