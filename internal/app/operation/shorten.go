package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

type shortener interface {
	Shorten(ctx context.Context, url string) (string, error)
}

type Shorten struct {
	Log     *logger.Logger
	BaseURL string
	Service interface {
		Shorten(ctx context.Context, url string) (string, error)
	}
}

func (o *Shorten) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	short, err := shorten(req.Context(), o.BaseURL, string(body), o.Service)
	if err != nil {
		o.Log.RequestError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(short))
}

type ShortenFromJSON struct {
	Log     *logger.Logger
	BaseURL string
	Service interface {
		Shorten(ctx context.Context, url string) (string, error)
	}
}

func (o *ShortenFromJSON) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var body struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	short, err := shorten(req.Context(), o.BaseURL, body.URL, o.Service)
	if err != nil {
		o.Log.RequestError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Result string `json:"result"`
	}{
		Result: short,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		o.Log.RequestError(req, fmt.Errorf("write response body: %w", err))
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
	id := s.IDGen.Next()

	encoded := s.Encoder.Encode(id)
	err := s.Storage.Add(ctx, id, encoded, url)
	if err != nil {
		return "", fmt.Errorf("shorten: failed to save ID: %v", err)
	}

	return encoded, nil
}
