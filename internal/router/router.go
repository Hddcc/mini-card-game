package router

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"mini-card-game/internal/cache"
	"mini-card-game/internal/config"
	"mini-card-game/internal/handler"
	"mini-card-game/internal/middleware"
	"mini-card-game/internal/mq"
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
	MQ     *mq.RabbitMQ
	Logger *zap.Logger
}

func frontendNoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" || strings.HasPrefix(path, "/static/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	}
}

func New(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(frontendNoCache())
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
	battleRepo := repository.NewBattleRepository(deps.DB)
	activityRepo := repository.NewActivityLotteryRepository(deps.DB)

	rewardService := service.NewRewardService(assetRepo, heroRepo)
	taskService := service.NewTaskService(deps.DB, taskRepo, rewardService)
	heroService := service.NewHeroService(heroRepo)
	gachaService := service.NewGachaService(deps.DB, assetRepo, gachaRepo, rewardService, taskService)
	teamService := service.NewTeamService(teamRepo, heroRepo, playerRepo)
	stageService := service.NewStageService(deps.DB, stageRepo, teamRepo, heroRepo, assetRepo, playerRepo, taskService)
	battleService := service.NewBattleService(deps.DB, battleRepo, stageRepo, teamRepo, heroRepo, assetRepo, playerRepo, taskService)
	activityCache := cache.NewRedisActivityCache(deps.Redis)
	activityService := service.NewActivityLotteryService(deps.DB, activityRepo, rewardService, activityCache, deps.MQ)
	service.StartActivityAwardConsumer(context.Background(), deps.MQ, activityService, deps.Logger)
	service.StartActivitySchedulers(context.Background(), activityService, deps.Logger)

	heroHandler := handler.NewHeroHandler(heroService)
	gachaHandler := handler.NewGachaHandler(gachaService)
	taskHandler := handler.NewTaskHandler(taskService)
	teamHandler := handler.NewTeamHandler(teamService)
	stageHandler := handler.NewStageHandler(stageService)
	battleHandler := handler.NewBattleHandler(battleService)
	activityHandler := handler.NewActivityLotteryHandler(activityService)

	frontendDist := deps.Config.FrontendDist
	if frontendDist == "" {
		frontendDist = "frontend/stitch"
	}

	if info, err := os.Stat(frontendDist); err == nil && info.IsDir() {
		r.Static("/static", frontendDist)
		indexHTML := filepath.Join(frontendDist, "mini_1", "code.html")
		fallbackPath := ""
		if _, err := os.Stat(indexHTML); err == nil {
			fallbackPath = indexHTML
		} else {
			_ = filepath.Walk(frontendDist, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				if strings.HasSuffix(strings.ToLower(info.Name()), ".html") {
					fallbackPath = path
					return errors.New("found")
				}
				return nil
			})
		}
		if fallbackPath != "" {
			r.GET("/", func(c *gin.Context) {
				c.File(fallbackPath)
			})
			r.NoRoute(func(c *gin.Context) {
				c.File(fallbackPath)
			})
		}
	}

	api := r.Group("/api/v1")
	loginRequired := api.Group("")
	loginRequired.Use(middleware.Auth(deps.Config.JWTSecret))
	loginRequired.GET("/player/profile", playerHandler.Profile)
	loginRequired.GET("/player/assets", playerHandler.Assets)
	loginRequired.GET("/heroes", heroHandler.List)
	loginRequired.GET("/gacha/state", gachaHandler.State)
	loginRequired.POST("/gacha/draw", gachaHandler.Draw)
	loginRequired.POST("/team/save", teamHandler.Save)
	loginRequired.GET("/team", teamHandler.Get)
	loginRequired.GET("/stages/progress", stageHandler.Progress)
	loginRequired.POST("/stage/fight", stageHandler.Fight)
	loginRequired.POST("/stage/battle/start", battleHandler.Start)
	loginRequired.POST("/stage/battle/action", battleHandler.Action)
	loginRequired.POST("/stage/battle/surrender", battleHandler.Surrender)
	loginRequired.GET("/tasks/daily", taskHandler.ListDaily)
	loginRequired.POST("/tasks/claim", taskHandler.Claim)
	loginRequired.GET("/activity/lottery/state", activityHandler.State)
	loginRequired.POST("/activity/lottery/draw", activityHandler.Draw)
	loginRequired.GET("/activity/lottery/records", activityHandler.Records)

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	return r
}
