package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisCli *redis.Client
func InitRedis() {
	RedisCli =  redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize:     10, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接数
		MaxRetries:   3,  // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时时间
		ReadTimeout:  3 * time.Second, // 读超时时间
		WriteTimeout: 3 * time.Second, // 写超时时间
	})

	ctx := context.Background()
    pong, err := RedisCli.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
	log.Printf("Redis successfully connected: %s", pong)
}

func CloseRedis() {
	if err := RedisCli.Close(); err != nil {
		log.Fatalf("Failed to close Redis connection: %v", err)
	}
	log.Println("Redis connection closed.")
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
	err := RedisCli.Set(ctx, shortCode, longURL, 0).Err()
	if err != nil {
		return err
	}
	return nil
}