package config

import (
	"flag"
	"os"
)

// Configuration of the app.
type Options struct {
	BaseURL            string `env:"BASE_URL"`
	DBConnectionString string `env:"DATABASE_DSN"`
	FileStoragePath    string `env:"FILE_STORAGE_PATH"`
	ServerAddress      string `env:"SERVER_ADDRESS"`
}

var opt Options

// Parse parses the options from command line or environment variables.
func Parse() {
	flag.StringVar(&opt.BaseURL, "b", "localhost:8080", "address of short url host")
	flag.StringVar(&opt.DBConnectionString, "d", "", "db connection string")
	flag.StringVar(&opt.FileStoragePath, "f", "/tmp/short-url-db.json", "name of file for storing short url")
	flag.StringVar(&opt.ServerAddress, "a", "localhost:8080", "host address")

	flag.Parse()

	if u := os.Getenv("BASE_URL"); u != "" {
		opt.BaseURL = u
	}

	if d := os.Getenv("DATABASE_DSN"); d != "" {
		opt.DBConnectionString = d
	}

	if f := os.Getenv("FILE_STORAGE_PATH"); f != "" {
		opt.FileStoragePath = f
	}

	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		opt.ServerAddress = a
	}
}

// Get returns options.
func Get() Options {
	return opt
}
