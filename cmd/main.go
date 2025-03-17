package main

import (
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"
)

func main() {
	router.Router()
	database.InitMysqlDB()
	cache.InitRedis()
	defer database.CloseMysqlDB()
	defer cache.CloseRedis()
}