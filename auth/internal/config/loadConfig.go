package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.AutomaticEnv()

	viper.BindEnv("http.port", "HTTP_PORT")

	viper.SetConfigFile("configs/auth/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	port := viper.GetString("http.port")
	fmt.Println("port==", port)
	if port != "" {
		viper.Set("http.port", port)
	}
}
