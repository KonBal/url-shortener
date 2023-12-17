package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/session"
)

// Represents operation to get urls of user.
type GetUserURLs struct {
	Log     *logger.Logger
	Service interface {
		GetUserURLs(ctx context.Context, userID string) ([]SavedURL, error)
	}
}

// ServeHTTP handles operation to get urls of user.
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

// Represents urls saved in the system.
type SavedURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserURLs returns URLs add by the user.
func (s ShortURLService) GetUserURLs(ctx context.Context, userID string) ([]SavedURL, error) {
	urls, err := s.Storage.GetURLsCreatedBy(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user urls: %w", err)
	}

	res := make([]SavedURL, 0, len(urls))
	for _, u := range urls {
		if !u.Deleted {
			res = append(res, SavedURL{
				ShortURL:    resolveURL(s.BaseURL, u.ShortURL),
				OriginalURL: u.OriginalURL,
			})
		}
	}

	return res, nil
}
