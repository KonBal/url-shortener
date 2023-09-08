package main

import (
	"log"
	"net/http"
	"os"

	"github.com/KonBal/url-shortener/internal/app/base62"
	"github.com/KonBal/url-shortener/internal/app/config"
	idgenerator "github.com/KonBal/url-shortener/internal/app/idgen"
	"github.com/KonBal/url-shortener/internal/app/operation"
	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.Parse()

	if err := run(); err != nil {
		log.Fatalf("main: unexpected error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	opt := config.Get()

	router := chi.NewRouter()

	c := base62.Encoder{}
	s := storage.NewInMemory()
	idgen := idgenerator.New()

	router.Post("/",
		operation.ShortenHandle(opt.BaseURL, operation.Shortener{
			Encoder: c,
			Storage: s,
			IDGen:   idgen,
		}))

	router.Get("/{short}",
		operation.ExpandHandle(operation.Expander{
			Decoder: c,
			Storage: s,
		}))

	return http.ListenAndServe(opt.ServerAddress, router)
}
