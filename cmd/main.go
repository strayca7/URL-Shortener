package main

import (
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
)

func main() {
	database.InitMysqlDB()
	cache.InitRedis()
	defer database.CloseMysqlDB()
	defer cache.CloseRedis()
}