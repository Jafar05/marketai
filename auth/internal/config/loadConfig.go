package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.AutomaticEnv()

	viper.BindEnv("http.port", "HTTP_PORT")

	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	// port := viper.GetString("http.port")
	// if port != "" {
	// 	viper.Set("http.port", port)
	// }
}
