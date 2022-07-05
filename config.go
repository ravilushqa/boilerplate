package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type config struct {
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPAddress string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0:10001"`
}

func newConfig() *config {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Printf("%+v\n", cfg)
	return &cfg
}
