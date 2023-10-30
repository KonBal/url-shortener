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
	"github.com/KonBal/url-shortener/internal/app/user"
	"github.com/KonBal/url-shortener/migrations"
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
		err = dbs.Bootstrap(migrations.SQLFiles)
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

	userStore := user.NewStore(randGen)
	authenticator := user.Authenticator{
		SecretKeyStore: user.NewKeyStore(func() []byte { return []byte("my_secret_key") }),
	}

	logged := LoggingHandler(log)
	compressed := ZipHandler()
	authorised := AuthHandler(authenticator)
	authenticated := AuthenticationHandler(authenticator, userStore)

	shortURLService := operation.ShortURLService{
		BaseURL:    opt.BaseURL,
		Encoder:    base62.Encoder{},
		Storage:    s,
		Uint64Rand: randGen,
	}
	deletionWorker := operation.NewDeletionWorker(s, log, 1024, 10)

	router.Method(http.MethodPost, "/",
		authenticated((compressed((&operation.Shorten{
			Log:     log,
			Service: shortURLService,
		})))))

	router.Method(http.MethodPost, "/api/shorten",
		authenticated((compressed((&operation.ShortenFromJSON{
			Log:     log,
			Service: shortURLService,
		})))))

	router.Method(http.MethodPost, "/api/shorten/batch",
		authenticated((compressed((&operation.ShortenBatch{
			Log:     log,
			Service: shortURLService,
		})))))

	router.Method(http.MethodGet, "/api/user/urls",
		authorised(logged(compressed(&operation.GetUserURLs{
			Log:     log,
			Service: shortURLService,
		}))))

	router.Method(http.MethodDelete, "/api/user/urls",
		authenticated(logged(&operation.Delete{
			Log:     log,
			Service: deletionWorker,
		})),
	)

	router.Method(http.MethodGet, "/{short}",
		authenticated((compressed((&operation.Expand{
			Log:     log,
			Service: shortURLService,
		})))))

	router.Method(http.MethodGet, "/ping", logged(&operation.Ping{Log: log, Storage: s}))

	return http.ListenAndServe(opt.ServerAddress, router)
}
