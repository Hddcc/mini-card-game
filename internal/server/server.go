package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mini-card-game/internal/cache"
	"mini-card-game/internal/config"
	"mini-card-game/internal/model"
	"mini-card-game/internal/pkg/logger"
	"mini-card-game/internal/router"
)

type Options struct {
	AllowMissingDatabase bool
}

func New(opts Options) (*gin.Engine, string, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, "", nil, err
	}

	zapLogger, err := logger.New(cfg.AppEnv)
	if err != nil {
		return nil, "", nil, err
	}
	cleanup := func() {
		_ = zapLogger.Sync()
	}

	var db *gorm.DB
	if cfg.MySQLDSN == "" {
		err := errors.New("MYSQL_DSN is required")
		if !opts.AllowMissingDatabase {
			cleanup()
			return nil, "", nil, err
		}
		zapLogger.Warn("database is not configured, serving static files only", zap.Error(err))
	} else {
		db, err = model.NewDB(cfg.MySQLDSN)
		if err != nil {
			if !opts.AllowMissingDatabase {
				cleanup()
				return nil, "", nil, err
			}
			zapLogger.Warn("init database failed, serving static files only", zap.Error(err))
		} else if err := model.EnsureSchema(db); err != nil {
			if !opts.AllowMissingDatabase {
				cleanup()
				return nil, "", nil, err
			}
			zapLogger.Warn("ensure database schema failed, serving static files only", zap.Error(err))
			db = nil
		}
	}

	redisClient, err := cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		zapLogger.Warn("init redis failed, continuing without redis", zap.Error(err))
		redisClient = nil
	}

	r := router.New(router.Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Logger: zapLogger,
	})

	return r, cfg.HTTPAddr, cleanup, nil
}
