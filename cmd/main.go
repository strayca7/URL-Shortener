package main

import (
	"os"
	_ "url-shortener/internal/config"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"

	"github.com/rs/zerolog/log"
)

func main() {
	pwd, _ := os.Getwd()
	log.Info().Str("pwd", pwd).Msg("Starting server")
	database.InitMysqlDB()
	defer database.CloseMysqlDB()
	router.Router()
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
