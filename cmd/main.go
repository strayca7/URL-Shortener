package main

import (
	// "url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"
)

func main() {
	router.Router()
	database.InitMysqlDB()
	defer database.CloseMysqlDB()
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
