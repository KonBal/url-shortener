package storage

import (
	"context"
	"errors"
)

// Storage.
type Storage interface {
	Add(ctx context.Context, url URLEntry, userID string) error
	AddMany(ctx context.Context, urls []URLEntry, userID string) error
	GetByShort(ctx context.Context, shortURL string) (*URLEntry, error)
	GetByOriginal(ctx context.Context, origURL string) (*URLEntry, error)
	GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error)
	MarkDeleted(ctx context.Context, urls ...EntryToDelete) error

	Ping(ctx context.Context) error
}

// Represents an entity of URL stored in storage.
type URLEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Deleted     bool   `json:"deleted,omitempty"`
}

// Entry to be marked deleted.
type EntryToDelete struct {
	ShortURL string
	UserID   string
}

// Error when entry no unique.
var ErrNotUnique = errors.New("not unique")

// Error when entry not found.
var ErrNotFound = errors.New("not found")
