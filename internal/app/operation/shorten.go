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
	"github.com/KonBal/url-shortener/internal/app/session"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

// Represents shorten operation.
type Shorten struct {
	Log     *logger.Logger
	Service interface {
		Shorten(ctx context.Context, userID string, url string) (string, error)
	}
}

// ServeHTTP handles shorten operation.
func (o *Shorten) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := req.Context()
	s := session.FromContext(ctx)

	status := http.StatusCreated

	short, err := o.Service.Shorten(ctx, s.UserID, string(body))
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

// Represents operation to shorten url recieved in JSON.
type ShortenFromJSON struct {
	Log     *logger.Logger
	Service interface {
		Shorten(ctx context.Context, userID string, url string) (string, error)
	}
}

// ServeHTTP handles operation to shorten url recieved in JSON.
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
	s := session.FromContext(ctx)

	status := http.StatusCreated

	short, err := o.Service.Shorten(ctx, s.UserID, body.URL)
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

// Represents operation to shorten multiple urls at once.
type ShortenBatch struct {
	Log     *logger.Logger
	Service interface {
		ShortenMany(ctx context.Context, userID string, orig []CorrelatedOrigURL) ([]CorrelatedShortURL, error)
	}
}

// ServeHTTP handles operation to shorten multiple urls at once.
func (o *ShortenBatch) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var urls []CorrelatedOrigURL

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := req.Context()
	s := session.FromContext(ctx)

	res, err := o.Service.ShortenMany(ctx, s.UserID, urls)
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

// Shorten computes a shortened URL for a given URL and saves both to the storage.
func (s ShortURLService) Shorten(ctx context.Context, userID string, url string) (string, error) {
	code := s.getEncoded()

	err := s.Storage.Add(ctx, storage.URLEntry{ShortURL: code, OriginalURL: url}, userID)
	switch {
	case errors.Is(err, storage.ErrNotUnique):
		sh, err := s.Storage.GetByOriginal(ctx, url)
		if err != nil {
			return "", err
		}

		return "", &notUniqueError{ShortURL: resolveURL(s.BaseURL, sh.ShortURL)}
	case err != nil:
		return "", fmt.Errorf("shorten: failed to save url: %w", err)
	}

	return resolveURL(s.BaseURL, code), nil
}

// Input type for original url.
type CorrelatedOrigURL struct {
	CorrelationID string `json:"correlation_id"`
	OrigURL       string `json:"original_url"`
}

// Result type for shortened url.
type CorrelatedShortURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenMany computes shortened URLs for given URLs and saves all to the storage.
func (s ShortURLService) ShortenMany(ctx context.Context, userID string, orig []CorrelatedOrigURL) ([]CorrelatedShortURL, error) {
	shorts := make([]CorrelatedShortURL, len(orig))
	entries := make([]storage.URLEntry, len(orig))

	for i, u := range orig {
		code := s.getEncoded()
		shorts[i] = CorrelatedShortURL{
			CorrelationID: u.CorrelationID, ShortURL: resolveURL(s.BaseURL, code),
		}
		entries[i] = storage.URLEntry{ShortURL: code, OriginalURL: u.OrigURL}
	}

	err := s.Storage.AddMany(ctx, entries, userID)
	if err != nil {
		return []CorrelatedShortURL{}, fmt.Errorf("shorten: failed to save urls: %w", err)
	}

	return shorts, nil
}

func (s ShortURLService) getEncoded() string {
	return s.Encoder.Encode(s.Uint64Rand.Next())
}

func resolveURL(baseURL string, short string) string {
	host := baseURL
	if !strings.Contains(host, "//") {
		host = "http://" + host
	}

	return fmt.Sprintf("%s/%s", host, short)
}
