package configs

import (
	. "godopi/internal/pkg/logger"

	"github.com/spf13/viper"
)

var config *viper.Viper

func init() {
	Logger().Info("Initializing config..")

	config = viper.New()

	setDefaults(config)

	config.AutomaticEnv()
	config.SetConfigType("env")
	config.SetConfigName("config")
	config.AddConfigPath("../config/")
	config.AddConfigPath("config")

	err := config.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error.
		} else {
			Logger().Fatal("Error on parsing configuration file!")
		}
	}
}

func Config() *viper.Viper {
	return config
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(SERVER_ADDRESS, "0.0.0.0:8080")
	config.SetDefault(REDIS_ADDRESS, "localhost:6379")
}
