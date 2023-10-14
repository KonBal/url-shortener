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

func (s InMemoryStorage) Add(ctx context.Context, uuid uint64, shortURL string, origURL string) error {
	lock.Lock()
	storage[shortURL] = origURL
	lock.Unlock()

	return nil
}

func (s InMemoryStorage) Get(ctx context.Context, shortURL string) (string, bool, error) {
	lock.RLock()
	v, ok := storage[shortURL]
	lock.RUnlock()

	return v, ok, nil
}
