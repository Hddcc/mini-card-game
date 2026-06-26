package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName          string
	AppEnv           string
	HTTPAddr         string
	FrontendDist     string
	MySQLDSN         string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	RabbitMQURL      string
	AwardExchange    string
	AwardQueue       string
	AwardRoutingKey  string
	JWTSecret        string
	JWTExpireSeconds int64
}

func Load() (*Config, error) {
	_ = godotenv.Load() // Load .env file if it exists
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, err
	}

	jwtExpire, err := strconv.ParseInt(getEnv("JWT_EXPIRE_SECONDS", "86400"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppName:          getEnv("APP_NAME", "mini-card-game"),
		AppEnv:           getEnv("APP_ENV", "local"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":5290"),
		FrontendDist:     getEnv("FRONTEND_DIST", "frontend/stitch"),
		MySQLDSN:         getEnv("MYSQL_DSN", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          redisDB,
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		AwardExchange:    getEnv("ACTIVITY_AWARD_EXCHANGE", "mini_xiyou.activity"),
		AwardQueue:       getEnv("ACTIVITY_AWARD_QUEUE", "mini_xiyou.activity.awards"),
		AwardRoutingKey:  getEnv("ACTIVITY_AWARD_ROUTING_KEY", "activity.award"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTExpireSeconds: jwtExpire,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
