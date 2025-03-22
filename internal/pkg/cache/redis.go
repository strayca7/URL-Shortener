package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var RedisCli *redis.Client

func InitRedis() {
	RedisCli = redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,              // 连接池大小
		MinIdleConns: 5,               // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时时间
		ReadTimeout:  3 * time.Second, // 读超时时间
		WriteTimeout: 3 * time.Second, // 写超时时间
	})

	ctx := context.Background()
	_, err := RedisCli.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	log.Debug().Msg("Successfully connected to Redis")
}

func CloseRedis() {
	if err := RedisCli.Close(); err != nil {
		log.Fatal().Err(err).Msg("Failed to close Redis connection")
	}
	log.Debug().Msg("Redis connection closed")
}

func GetURL(shortCode string) (string, error) {
	ctx := context.Background()
	longURL, err := RedisCli.Get(ctx, shortCode).Result()
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func SetURL(shortCode, longURL string) error {
	ctx := context.Background()
	err := RedisCli.Set(ctx, shortCode, longURL, 90*24*time.Duration(time.Hour)).Err()
	if err != nil {
		return err
	}
	return nil
}
