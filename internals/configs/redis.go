package configs

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, string, error) {
	rdbUser := os.Getenv("REDISUSER")
	rdbPass := os.Getenv("REDISPASS")
	rdbHost := os.Getenv("REDISHOST")
	rdbPort := os.Getenv("REDISPORT")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdbHost, rdbPort),
		Username: rdbUser,
		Password: rdbPass,
		DB:       0,
	})
	return rdb, rdbUser, nil
}

var ctx = context.Background()

func NewRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("UPSTASH_REDIS_URL"),
		Password: os.Getenv("UPSTASH_REDIS_PASSWORD"),
		DB:       0,
	})
}
