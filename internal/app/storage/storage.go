package storage

import "context"

type Storage interface {
	Add(ctx context.Context, uuid uint64, shortURL string, origURL string) error
	Get(ctx context.Context, shortURL string) (string, bool, error)
}
