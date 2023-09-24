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
	case err != nil:
		o.Log.RequestError(req, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

type Expander struct {
	Storage storage.Storage
}

func (r Expander) Expand(ctx context.Context, shortened string) (string, error) {
	url, ok, err := r.Storage.Get(shortened)
	if err != nil {
		return "", fmt.Errorf("expand: failed to get original URL: %w", err)
	}
	if !ok {
		return "", notFoundError(fmt.Sprintf("original URL not found for shortened %s", shortened))
	}

	return url, nil
}
