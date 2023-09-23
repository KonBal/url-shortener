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
	"go.uber.org/zap"
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

	baseLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer baseLogger.Sync()

	logger := baseLogger.Sugar()

	logged := LoggingHandler(logger)
	compressed := ZipHandler()

	c := base62.Encoder{}
	s := storage.NewInMemory()
	idgen := idgenerator.New()

	router.Post("/",
		logged(compressed(operation.ShortenHandle(logger, opt.BaseURL, operation.Shortener{
			Encoder: c,
			Storage: s,
			IDGen:   idgen,
		}))))

	router.Post("/api/shorten",
		logged(compressed(operation.ShortenFromJSONHandle(logger, opt.BaseURL, operation.Shortener{
			Encoder: c,
			Storage: s,
			IDGen:   idgen,
		}))))

	router.Get("/{short}",
		logged(compressed(operation.ExpandHandle(logger, operation.Expander{
			Decoder: c,
			Storage: s,
		}))))

	return http.ListenAndServe(opt.ServerAddress, router)
}
