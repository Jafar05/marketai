package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() *Config {
	viper.AutomaticEnv()
	viper.BindEnv("http.port", "HTTP_PORT")

	viper.SetConfigFile("configs/auth/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Ошибка Unmarshal конфига: %v", err)
	}

	port := viper.GetString("http.port")
	fmt.Println("port==", port)
	if port != "" {
		cfg.Http.Port = port
	}

	return &cfg
}
