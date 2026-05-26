package router

import (
	"mini-card-game/internal/config"

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

	api := r.Group("/api/v1")
	_ = api

	return r
}
