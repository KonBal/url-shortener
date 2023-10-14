package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/KonBal/url-shortener/internal/app/base62"
	"github.com/KonBal/url-shortener/internal/app/config"
	idgenerator "github.com/KonBal/url-shortener/internal/app/idgen"
	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/operation"
	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	config.Parse()

	baseLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("main: failed to initialize custom logger: %v", err)
		return
	}
	defer baseLogger.Sync()

	customLog := logger.NewLogger(baseLogger)

	if err := run(customLog); err != nil {
		customLog.Fatalf("main: unexpected error: %w", err)
		os.Exit(1)
	}
}

func run(log *logger.Logger) error {
	opt := config.Get()

	router := chi.NewRouter()

	db, err := sql.Open("pgx", opt.DBConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	logged := LoggingHandler(log)
	compressed := ZipHandler()

	var s storage.Storage

	if opt.DBConnectionString != "" {
		dbs, err := storage.NewDBStorage(opt.DBConnectionString)
		if err != nil {
			return err
		}
		defer dbs.Close()

		s = dbs
	} else if opt.FileStoragePath != "" {
		fs, err := storage.NewFileStorage(opt.FileStoragePath)
		if err != nil {
			return err
		}
		defer fs.Close()

		s = fs
	} else {
		s = storage.NewInMemory()
	}

	shortener := operation.Shortener{
		Encoder: base62.Encoder{},
		Storage: s,
		IDGen:   idgenerator.New(),
	}

	expander := operation.Expander{Storage: s}

	router.Method(http.MethodPost, "/",
		logged(compressed((&operation.Shorten{
			Log:     log,
			BaseURL: opt.BaseURL,
			Service: shortener,
		}))))

	router.Method(http.MethodPost, "/api/shorten",
		logged(compressed((&operation.ShortenFromJSON{
			Log:     log,
			BaseURL: opt.BaseURL,
			Service: shortener,
		}))))

	router.Method(http.MethodGet, "/{short}",
		logged(compressed((&operation.Expand{
			Log:     log,
			Service: expander,
		}))))

	router.Method(http.MethodGet, "/ping", logged(&operation.Ping{Log: log, DB: db}))

	return http.ListenAndServe(opt.ServerAddress, router)
}
