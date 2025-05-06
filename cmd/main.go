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
	log.Info().Str("wd", wd).Msg("** Starting server **")

	if err := database.InitMysqlDB(); err != nil {
		log.Fatal().Msg("Failed to connect to MySQL")
		return
	}
	defer func() {
		if err := database.CloseMysqlDB(); err != nil {
			log.Err(err).Msg("Error closing MySQL connection")
		}
	}()

	router.Router()
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
