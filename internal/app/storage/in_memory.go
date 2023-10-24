package storage

import (
	"context"
	"sync"
)

type InMemoryStorage map[string]inMemoryEntry

type inMemoryEntry struct {
	OriginalURL string
	CreatedBy   string
}

var storage InMemoryStorage
var lock *sync.RWMutex

func NewInMemory() InMemoryStorage {
	storage = make(map[string]inMemoryEntry)
	lock = &sync.RWMutex{}
	return storage
}

func (s InMemoryStorage) Add(ctx context.Context, u URLEntry, userID string) error {
	lock.Lock()
	storage[u.ShortURL] = inMemoryEntry{OriginalURL: u.OriginalURL, CreatedBy: userID}
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) AddMany(ctx context.Context, urls []URLEntry, userID string) error {
	lock.Lock()
	for _, u := range urls {
		storage[u.ShortURL] = inMemoryEntry{OriginalURL: u.OriginalURL, CreatedBy: userID}
	}
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	lock.RLock()
	v, ok := storage[shortURL]
	lock.RUnlock()

	if !ok {
		return "", ErrNotFound
	}

	return v.OriginalURL, nil
}

func (s InMemoryStorage) GetShort(ctx context.Context, origURL string) (string, error) {
	lock.RLock()
	for k, v := range storage {
		if v.OriginalURL == origURL {
			return k, nil
		}
	}
	lock.RUnlock()

	return "", ErrNotFound
}

func (s InMemoryStorage) GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error) {
	var urls []URLEntry

	lock.RLock()
	for k, v := range storage {
		if v.CreatedBy == userID {
			urls = append(urls, URLEntry{ShortURL: k, OriginalURL: v.OriginalURL})
		}
	}
	lock.RUnlock()

	return urls, nil
}

func (s InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
