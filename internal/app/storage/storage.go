package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Add(ctx context.Context, url URLEntry, userID string) error
	AddMany(ctx context.Context, urls []URLEntry, userID string) error
	GetByShort(ctx context.Context, shortURL string) (*URLEntry, error)
	GetByOriginal(ctx context.Context, origURL string) (*URLEntry, error)
	GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error)
	MarkDeleted(ctx context.Context, urls ...EntryToDelete) error

	Ping(ctx context.Context) error
}

type URLEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Deleted     bool   `json:"deleted,omitempty"`
}

type EntryToDelete struct {
	ShortURL string
	UserID   string
}

var ErrNotUnique = errors.New("not unique")
var ErrNotFound = errors.New("not found")
