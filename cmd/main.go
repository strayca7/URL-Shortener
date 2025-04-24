package main

import (
	"os"
	_ "url-shortener/internal/config"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"

	"github.com/rs/zerolog/log"
)

func main() {
	wd, _ := os.Getwd()
	log.Info().Str("wd", wd).Msg("Starting server")
	err := database.InitMysqlDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}
	defer database.CloseMysqlDB()
	router.Router()
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
