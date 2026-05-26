package cache

import (
	"github.com/go-redis/redis"
)

func NewRedis(addr string, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}
	return client, nil
}
