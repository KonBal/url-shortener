package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func ExpandHandle(logger interface {
	Errorf(template string, args ...interface{})
}, e interface {
	Expand(ctx context.Context, shortened string) (string, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortened := chi.URLParam(r, "short")
		ctx := r.Context()
		url, err := e.Expand(ctx, shortened)

		switch {
		case errors.Is(err, ErrNotFound):
			logError(logger, r, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		case err != nil:
			logError(logger, r, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

type Expander struct {
	Decoder interface {
		Decode(v string) (uint64, error)
	}
	Storage storage.Storage
}

func (r Expander) Expand(ctx context.Context, shortened string) (string, error) {
	ID, err := r.Decoder.Decode(shortened)
	if err != nil {
		return "", fmt.Errorf("expand: failed to decode: %w", err)
	}

	url, ok := r.Storage.Get(ID)
	if !ok {
		return "", notFoundError(fmt.Sprintf("original URL not found for shortened %s", shortened))
	}

	return url, nil
}
