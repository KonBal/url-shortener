package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Shorten struct {
	Log     *logger.Logger
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

	ctx := req.Context()
	status := http.StatusCreated

	short, err := o.Service.Shorten(ctx, string(body))
	if err != nil {
		var errUnique *notUniqueError
		if errors.As(err, &errUnique) {
			status = http.StatusConflict
			short = errUnique.ShortURL
		} else {
			o.Log.RequestError(req, err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(short))
}

type ShortenFromJSON struct {
	Log     *logger.Logger
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

	ctx := req.Context()
	status := http.StatusCreated

	short, err := o.Service.Shorten(ctx, body.URL)
	if err != nil {
		var errUnique *notUniqueError
		if errors.As(err, &errUnique) {
			status = http.StatusConflict
			short = errUnique.ShortURL
		} else {
			o.Log.RequestError(req, err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}
	}

	resp := struct {
		Result string `json:"result"`
	}{
		Result: short,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		o.Log.RequestError(req, fmt.Errorf("write response body: %w", err))
	}
}

type ShortenBatch struct {
	Log     *logger.Logger
	Service interface {
		ShortenMany(ctx context.Context, orig []CorrelatedOrigURL) ([]CorrelatedShortURL, error)
	}
}

func (o *ShortenBatch) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var urls []CorrelatedOrigURL

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := req.Context()

	res, err := o.Service.ShortenMany(ctx, urls)
	if err != nil {
		o.Log.RequestError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		o.Log.RequestError(req, fmt.Errorf("write response body: %w", err))
	}
}

type Shortener struct {
	BaseURL string
	Encoder interface {
		Encode(v uint64) string
	}
	Storage    storage.Storage
	Uint64Rand interface {
		Next() uint64
	}
}

func (s Shortener) Shorten(ctx context.Context, url string) (string, error) {
	code := s.getEncoded()

	err := s.Storage.Add(ctx, storage.URLEntry{ShortURL: code, OriginalURL: url})
	switch {
	case errors.Is(err, storage.ErrNotUnique):
		sh, err := s.Storage.GetShort(ctx, url)
		if err != nil {
			return "", err
		}

		return "", &notUniqueError{ShortURL: resolveURL(s.BaseURL, sh)}
	case err != nil:
		return "", fmt.Errorf("shorten: failed to save url: %w", err)
	}

	return resolveURL(s.BaseURL, code), nil
}

type CorrelatedOrigURL struct {
	CorrelationID string `json:"correlation_id"`
	OrigURL       string `json:"original_url"`
}
type CorrelatedShortURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s Shortener) ShortenMany(ctx context.Context, orig []CorrelatedOrigURL) ([]CorrelatedShortURL, error) {
	shorts := make([]CorrelatedShortURL, len(orig))
	entries := make([]storage.URLEntry, len(orig))

	for i, u := range orig {
		code := s.getEncoded()
		shorts[i] = CorrelatedShortURL{
			CorrelationID: u.CorrelationID, ShortURL: resolveURL(s.BaseURL, code),
		}
		entries[i] = storage.URLEntry{ShortURL: code, OriginalURL: u.OrigURL}
	}

	err := s.Storage.AddMany(ctx, entries)
	if err != nil {
		return []CorrelatedShortURL{}, fmt.Errorf("shorten: failed to save urls: %w", err)
	}

	return shorts, nil
}

func (s Shortener) getEncoded() string {
	return s.Encoder.Encode(s.Uint64Rand.Next())
}

func resolveURL(baseURL string, short string) string {
	host := baseURL
	if !strings.Contains(host, "//") {
		host = "http://" + host
	}

	return fmt.Sprintf("%s/%s", host, short)
}
