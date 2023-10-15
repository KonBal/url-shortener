package storage

import (
	"context"
	"sync"
)

type InMemoryStorage map[string]string

var storage InMemoryStorage
var lock *sync.RWMutex

func NewInMemory() InMemoryStorage {
	storage = make(map[string]string)
	lock = &sync.RWMutex{}
	return storage
}

func (s InMemoryStorage) Add(ctx context.Context, u URLEntry) error {
	lock.Lock()
	storage[u.ShortURL] = u.OriginalURL
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) AddMany(ctx context.Context, urls []URLEntry) error {
	lock.Lock()
	for _, u := range urls {
		storage[u.ShortURL] = u.OriginalURL
	}
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) GetOriginal(ctx context.Context, shortURL string) (string, bool, error) {
	lock.RLock()
	v, ok := storage[shortURL]
	lock.RUnlock()

	return v, ok, nil
}

func (s InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
