package main

import (
	"marketai/cards/internal/ports"

	"go.uber.org/fx"
)

func main() {
	fx.New(ports.App()).Run()
}
