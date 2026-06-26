package main

import (
	"go.uber.org/zap"
	"log"
	"mini-card-game/internal/cache"
	"mini-card-game/internal/config"
	"mini-card-game/internal/model"
	"mini-card-game/internal/mq"
	"mini-card-game/internal/pkg/logger"
	"mini-card-game/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	zapLogger, err := logger.New(cfg.AppEnv)
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer zapLogger.Sync()

	db, err := model.NewDB(cfg.MySQLDSN)
	if err != nil {
		zapLogger.Fatal("init database failed", zap.Error(err))
	}
	if err := model.EnsureSchema(db); err != nil {
		zapLogger.Fatal("ensure database schema failed", zap.Error(err))
	}

	redisClient, err := cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		zapLogger.Warn("init redis failed, continuing without redis", zap.Error(err))
		redisClient = nil
	}

	rabbitClient, err := mq.NewRabbitMQ(mq.Config{
		URL:        cfg.RabbitMQURL,
		Exchange:   cfg.AwardExchange,
		Queue:      cfg.AwardQueue,
		RoutingKey: cfg.AwardRoutingKey,
	})
	if err != nil {
		zapLogger.Warn("init rabbitmq failed, continuing with local message retry", zap.Error(err))
		rabbitClient = nil
	} else {
		defer rabbitClient.Close()
	}

	r := router.New(router.Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		MQ:     rabbitClient,
		Logger: zapLogger,
	})

	if err := r.Run(cfg.HTTPAddr); err != nil {
		zapLogger.Fatal("server failed to start")
	}
}
