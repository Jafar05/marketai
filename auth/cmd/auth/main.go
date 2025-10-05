package main

import (
	"marketai/auth/internal/config"
	"marketai/auth/internal/ports"

	"go.uber.org/fx"
)

func main() {
	config.LoadConfig()
	fx.New(ports.App()).Run()
}
