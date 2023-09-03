package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Expand struct {
	Service interface {
		Expand(ctx context.Context, shortend string) (string, error)
	}
}

func (o Expand) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		logError(req, fmt.Errorf("expected method GET, got %v", req.Method))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	encoded, err := retrieveID(req.URL.Path)
	if err != nil {
		logError(req, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	url, err := o.Service.Expand(ctx, encoded)
	switch {
	case errors.Is(err, ErrNotFound):
		logError(req, err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case err != nil:
		logError(req, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func retrieveID(url string) (string, error) {
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}

	sp := strings.Split(url, "/")
	if len(sp) < 2 {
		return "", fmt.Errorf("id is not found in url")
	}

	return sp[1], nil
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
		return "", err
	}

	url, ok := r.Storage.Get(ID)
	if !ok {
		return "", notFoundError(fmt.Sprintf("original URL not found for shortened %s", shortened))
	}

	return url, nil
}
