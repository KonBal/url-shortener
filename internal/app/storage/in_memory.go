package storage

import (
	"context"
	"sync"
)

type InMemoryStorage map[string]inMemoryEntry

type inMemoryEntry struct {
	OriginalURL string
	CreatedBy   string
	Deleted     bool
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

func (s InMemoryStorage) GetByShort(ctx context.Context, shortURL string) (*URLEntry, error) {
	lock.RLock()
	v, ok := storage[shortURL]
	lock.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	return &URLEntry{ShortURL: shortURL, OriginalURL: v.OriginalURL, Deleted: v.Deleted}, nil
}

func (s InMemoryStorage) GetByOriginal(ctx context.Context, origURL string) (*URLEntry, error) {
	lock.RLock()
	for k, v := range storage {
		if v.OriginalURL == origURL {
			return &URLEntry{ShortURL: k, OriginalURL: origURL, Deleted: v.Deleted}, nil
		}
	}
	lock.RUnlock()

	return nil, ErrNotFound
}

func (s InMemoryStorage) GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error) {
	var urls []URLEntry

	lock.RLock()
	for k, v := range storage {
		if v.CreatedBy == userID {
			urls = append(urls, URLEntry{ShortURL: k, OriginalURL: v.OriginalURL, Deleted: v.Deleted})
		}
	}
	lock.RUnlock()

	return urls, nil
}

func (s InMemoryStorage) MarkDeleted(ctx context.Context, urls ...EntryToDelete) error {
	lock.Lock()
	for _, u := range urls {
		entry, ok := s[u.ShortURL]
		if ok && entry.CreatedBy == u.UserID {
			entry.Deleted = true
			s[u.ShortURL] = entry
		}
	}
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
