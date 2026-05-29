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
	assetRepo := repository.NewAssetRepository(deps.DB)
	heroRepo := repository.NewHeroRepository(deps.DB)
	gachaRepo := repository.NewGachaRepository(deps.DB)
	taskRepo := repository.NewTaskRepository(deps.DB)
	teamRepo := repository.NewTeamRepository(deps.DB)
	stageRepo := repository.NewStageRepository(deps.DB)

	rewardService := service.NewRewardService(assetRepo, heroRepo)
	taskService := service.NewTaskService(deps.DB, taskRepo, rewardService)
	heroService := service.NewHeroService(heroRepo)
	gachaService := service.NewGachaService(deps.DB, assetRepo, gachaRepo, rewardService, taskService)
	teamService := service.NewTeamService(teamRepo, heroRepo, playerRepo)
	stageService := service.NewStageService(deps.DB, stageRepo, teamRepo, heroRepo, assetRepo, playerRepo, taskService)

	heroHandler := handler.NewHeroHandler(heroService)
	gachaHandler := handler.NewGachaHandler(gachaService)
	taskHandler := handler.NewTaskHandler(taskService)
	teamHandler := handler.NewTeamHandler(teamService)
	stageHandler := handler.NewStageHandler(stageService)

	api := r.Group("/api/v1")
	loginRequired := api.Group("")
	loginRequired.Use(middleware.Auth(deps.Config.JWTSecret))
	loginRequired.GET("/player/profile", playerHandler.Profile)
	loginRequired.GET("/player/assets", playerHandler.Assets)
	loginRequired.GET("/heroes", heroHandler.List)
	loginRequired.POST("/gacha/draw", gachaHandler.Draw)
	loginRequired.POST("/team/save", teamHandler.Save)
	loginRequired.GET("/team", teamHandler.Get)
	loginRequired.POST("/stage/fight", stageHandler.Fight)
	loginRequired.GET("/tasks/daily", taskHandler.ListDaily)
	loginRequired.POST("/tasks/claim", taskHandler.Claim)

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	return r
}
