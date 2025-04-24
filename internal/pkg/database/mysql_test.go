package database

import (
	"errors"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file")
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}

	var multiWriter zerolog.LevelWriter = zerolog.MultiLevelWriter(consoleWriter)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger()

	log.Debug().Msg("Init logger")
	log.Info().Err(errors.New("test error")).Msg("error")
}

func TestConnectDB(t *testing.T) {
	// Test MySQL connection
	InitMysqlDB()
	defer CloseMysqlDB()
}
