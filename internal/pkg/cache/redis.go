package cache

import (
	"context"
	"time"
	"url-shortener/internal/pkg/database"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var rDB *redis.Client

func InitRedis() {
	log.Debug().Msg("** Start init redis **")
	rDB = redis.NewClient(&redis.Options{
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
	_, err := rDB.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	log.Debug().Msg("Successfully connected to Redis")
}

func CloseRedis() {
	if err := rDB.Close(); err != nil {
		log.Fatal().Err(err).Msg("Failed to close Redis connection")
	}
	log.Debug().Msg("Redis connection closed")
}

func GetURL(shortCode string) (string, error) {
	ctx := context.Background()
	longURL, err := rDB.Get(ctx, shortCode).Result()
	if err != nil {
		return "", err
	}
	return longURL, nil
}

// 存储 ShortURL 到 Redis
func SaveShortURL(url database.UserShortURL) error {
	ctx := context.Background()
	// 使用 Hash 存储短链接元数据
	err := rDB.HSet(ctx, "shorturl:"+url.ShortCode,
		"original_url", url.OriginalURL,
		"expire_at", url.ExpireAt.Format(time.RFC3339),
		"user_id", url.UserID,
	).Err()
	if err != nil {
		return err
	}

	// 设置过期时间（若需）
	if !url.ExpireAt.IsZero() {
		rDB.ExpireAt(ctx, "shorturl:"+url.ShortCode, url.ExpireAt)
	}
	return nil
}
