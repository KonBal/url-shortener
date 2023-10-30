package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type Expand struct {
	Log     *logger.Logger
	Service interface {
		Expand(ctx context.Context, shortened string) (string, error)
	}
}

func (o *Expand) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	shortened := chi.URLParam(req, "short")
	ctx := req.Context()
	url, err := o.Service.Expand(ctx, shortened)

	switch {
	case errors.Is(err, ErrNotFound):
		o.Log.RequestError(req, err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	case errors.Is(err, ErrDeleted):
		o.Log.RequestError(req, err)
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		return
	case err != nil:
		o.Log.RequestError(req, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s ShortURLService) Expand(ctx context.Context, shortened string) (string, error) {
	u, err := s.Storage.GetByShort(ctx, shortened)
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return "", notFoundError(fmt.Sprintf("original URL not found for shortened %s", shortened))
	case err != nil:
		return "", fmt.Errorf("expand: failed to get original URL: %w", err)
	}

	if u.Deleted {
		return "", deletedError(fmt.Sprintf("url for shortened %s is already deleted", shortened))
	}

	return u.OriginalURL, nil
}
