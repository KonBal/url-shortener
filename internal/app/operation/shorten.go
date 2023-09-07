package operation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Shorten struct {
	Service interface {
		Shorten(ctx context.Context, url string) (string, error)
	}
}

func ShortenHandle(shortURLHost string, s interface {
	Shorten(ctx context.Context, url string) (string, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logError(r, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx := r.Context()

		url := string(body)
		short, err := s.Shorten(ctx, url)
		if err != nil {
			logError(r, err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}

		sHost := shortURLHost
		if !strings.Contains(sHost, "//") {
			sHost = "http://" + sHost
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("%s/%s", sHost, short)))
	}
}

type Shortener struct {
	Encoder interface {
		Encode(v uint64) string
	}
	Storage storage.Storage
	IDGen   interface {
		Next() uint64
	}
}

func (s Shortener) Shorten(ctx context.Context, url string) (string, error) {
	id := s.findFreeID()

	err := s.Storage.Set(id, url)
	if err != nil {
		return "", err
	}

	encoded := s.Encoder.Encode(id)

	return encoded, nil
}

func (s Shortener) findFreeID() uint64 {
	var ID uint64
	for used := true; used; used = s.Storage.HasKey(ID) {
		ID = s.IDGen.Next()
	}
	return ID
}
