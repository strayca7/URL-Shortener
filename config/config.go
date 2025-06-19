package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	TestMode bool
)

func init() {
	// // 注意运行时文件目录，确保每次启动都从根目录运行
	// if err := os.Chdir("./"); err != nil {
	// 	panic(err)
	// }

	initLogger()
	initViper()

	TestMode = viper.GetBool("test_mode")

	log.Debug().Msg("Init finish")
}

func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// 将配置文件从 /app 修改至 /etc
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err)
	}
	log.Debug().Msg("Init viper")
}

func initLogger() {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}
	// consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	fileWriter := &lumberjack.Logger{
		Filename:   "./log/app_" + time.Now().String() + ".log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}
	// fileLogger := zerolog.New(fileWriter).With().Timestamp().Logger()

	var multiWriter zerolog.LevelWriter = zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger()

	log.Debug().Msg("Init logger")
}
