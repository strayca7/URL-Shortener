package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func init() {
	if err := os.Chdir("../.."); err != nil {
        panic(err)
    }
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./internal/config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
func Host() string {
	return "localhost"
}
