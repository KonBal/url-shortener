package operation

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Shorten struct {
	Service interface {
		Shorten(ctx context.Context, url string) (string, error)
	}
}

func (o Shorten) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		logError(req, fmt.Errorf("expected method POST, got %v", req.Method))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logError(req, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := req.Context()

	url := string(body)
	short, err := o.Service.Shorten(ctx, url)
	if err != nil {
		logError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(short))
}

type Shortener struct {
	Encoder interface {
		Encode(v uint64) string
	}
	Storage      storage.Storage
	ShortURLHost interface {
		Address() string
	}
	IDGen interface {
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

	return fmt.Sprintf("%s/%s", s.ShortURLHost.Address(), encoded), nil
}

func (s Shortener) findFreeID() uint64 {
	var ID uint64
	for used := true; used; used = s.Storage.HasKey(ID) {
		ID = s.IDGen.Next()
	}
	return ID
}
