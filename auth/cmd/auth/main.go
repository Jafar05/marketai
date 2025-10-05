package main

import (
	"marketai/auth/internal/ports"

	"go.uber.org/fx"
)

func main() {
	fx.New(ports.App()).Run()
}
