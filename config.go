package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type config struct {
	Env            string `env:"ENV" envDefault:"development"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPAddress    string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0:8081"`
	AppHTTPAddress string `env:"APP_HTTP_ADDRESS" envDefault:"0.0.0.0:8080"`
}

func newConfig() *config {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Printf("%+v\n", cfg)
	return &cfg
}
