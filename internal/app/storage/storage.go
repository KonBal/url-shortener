package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Add(ctx context.Context, url URLEntry) error
	AddMany(ctx context.Context, urls []URLEntry) error
	GetOriginal(ctx context.Context, shortURL string) (string, error)
	GetShort(ctx context.Context, origURL string) (string, error)

	Ping(ctx context.Context) error
}

type URLEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var ErrNotUnique = errors.New("not unique")
var ErrNotFound = errors.New("not found")
