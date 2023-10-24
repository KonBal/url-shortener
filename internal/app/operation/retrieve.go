package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/session"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

type GetUserURLs struct {
	Log     *logger.Logger
	Service interface {
		GetUserURLs(ctx context.Context, userID string) ([]SavedURL, error)
	}
}

func (o *GetUserURLs) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	s := session.FromContext(ctx)

	resp, err := o.Service.GetUserURLs(ctx, s.UserID)
	if err != nil {
		o.Log.RequestError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
		return
	}

	if len(resp) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		o.Log.RequestError(req, fmt.Errorf("write response body: %w", err))
	}
}

type SavedURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type RetrieveService struct {
	BaseURL string
	Storage interface {
		GetURLsCreatedBy(ctx context.Context, userID string) ([]storage.URLEntry, error)
	}
}

func (s RetrieveService) GetUserURLs(ctx context.Context, userID string) ([]SavedURL, error) {
	urls, err := s.Storage.GetURLsCreatedBy(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user urls: %w", err)
	}

	var res []SavedURL
	for _, u := range urls {
		res = append(res, SavedURL{
			ShortURL:    resolveURL(s.BaseURL, u.ShortURL),
			OriginalURL: u.OriginalURL,
		})
	}

	return res, nil
}
