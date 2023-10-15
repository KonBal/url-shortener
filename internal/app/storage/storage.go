package storage

import "context"

type Storage interface {
	Add(ctx context.Context, url URLEntry) error
	AddMany(ctx context.Context, urls []URLEntry) error
	GetOriginal(ctx context.Context, shortURL string) (string, bool, error)

	Ping(ctx context.Context) error
}

type URLEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
