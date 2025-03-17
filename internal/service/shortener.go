package service

import (
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// 集成 MySQL,Redis 存储URL
func SaveURL(url, shortCode string) error {
    if err := database.SaveURL(shortCode, url); err != nil {
        return err
    }
    if err := cache.SetURL(shortCode, url); err != nil {
        return err
    }
    return nil
}