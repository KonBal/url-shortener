package main

import (
	"database/sql"
	"fmt"
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

	logged := LoggingHandler(log)
	compressed := ZipHandler()

	randGen := idgenerator.New()

	var s storage.Storage

	switch {
	case opt.DBConnectionString != "":
		db, err := sql.Open("pgx", opt.DBConnectionString)
		if err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		} else if err = db.Ping(); err != nil {
			return err
		}
		defer db.Close()

		dbs := storage.NewDBStorage(db)
		err = dbs.Bootstrap()
		if err != nil {
			return err
		}

		s = dbs
	case opt.FileStoragePath != "":
		fs, err := storage.NewFileStorage(opt.FileStoragePath, randGen)
		if err != nil {
			return err
		}
		defer fs.Close()

		s = fs
	default:
		s = storage.NewInMemory()
	}

	shortener := operation.Shortener{
		BaseURL:    opt.BaseURL,
		Encoder:    base62.Encoder{},
		Storage:    s,
		Uint64Rand: randGen,
	}

	expander := operation.Expander{Storage: s}

	router.Method(http.MethodPost, "/",
		logged(compressed((&operation.Shorten{
			Log:     log,
			Service: shortener,
		}))))

	router.Method(http.MethodPost, "/api/shorten",
		logged(compressed((&operation.ShortenFromJSON{
			Log:     log,
			Service: shortener,
		}))))

	router.Method(http.MethodPost, "/api/shorten/batch",
		logged(compressed((&operation.ShortenBatch{
			Log:     log,
			Service: shortener,
		}))))

	router.Method(http.MethodGet, "/{short}",
		logged(compressed((&operation.Expand{
			Log:     log,
			Service: expander,
		}))))

	router.Method(http.MethodGet, "/ping", logged(&operation.Ping{Log: log, Storage: s}))

	return http.ListenAndServe(opt.ServerAddress, router)
}
