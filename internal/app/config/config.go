package config

import (
	"flag"
	"os"
)

type Options struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var opt Options

func Parse() {
	flag.StringVar(&opt.ServerAddress, "a", "localhost:8080", "host address")
	flag.StringVar(&opt.BaseURL, "b", "localhost:8080", "address of short url host")

	flag.Parse()

	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		opt.ServerAddress = a
	}

	if u := os.Getenv("BASE_URL"); u != "" {
		opt.BaseURL = u
	}
}

func Get() Options {
	return opt
}
