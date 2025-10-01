package main

import (
	"go.uber.org/fx"
	"marketai/auth/internal/ports"
)

func main() {
	fx.New(ports.App()).Run()
}
