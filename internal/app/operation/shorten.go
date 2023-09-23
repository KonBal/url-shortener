package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KonBal/url-shortener/internal/app/storage"
)

type shortener interface {
	Shorten(ctx context.Context, url string) (string, error)
}

func ShortenHandle(logger interface {
	Errorf(template string, args ...interface{})
}, baseURL string, s shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logError(logger, r, fmt.Errorf("read request body: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		short, err := shorten(r.Context(), baseURL, string(body), s)
		if err != nil {
			logError(logger, r, err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(short))
	}
}

func ShortenFromJSONHandle(logger interface {
	Errorf(template string, args ...interface{})
}, baseURL string, s shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			URL string `json:"url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			logError(logger, r, fmt.Errorf("read request body: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		short, err := shorten(r.Context(), baseURL, body.URL, s)
		if err != nil {
			logError(logger, r, err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
		}

		resp := struct {
			Result string `json:"result"`
		}{
			Result: short,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logError(logger, r, fmt.Errorf("write response body: %w", err))
		}
	}
}

func shorten(ctx context.Context, baseURL, url string, s shortener) (string, error) {
	short, err := s.Shorten(ctx, url)
	if err != nil {
		return "", err
	}

	host := baseURL
	if !strings.Contains(host, "//") {
		host = "http://" + host
	}

	return fmt.Sprintf("%s/%s", host, short), nil
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
		return "", fmt.Errorf("shorten: failed to save ID: %v", err)
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
