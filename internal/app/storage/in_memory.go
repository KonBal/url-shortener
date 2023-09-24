package storage

import "sync"

type inMemoryStorage map[string]string

var storage inMemoryStorage
var lock *sync.RWMutex

func NewInMemory() Storage {
	storage = make(map[string]string)
	lock = &sync.RWMutex{}
	return storage
}

func (s inMemoryStorage) Add(uuid uint64, shortURL string, origURL string) error {
	lock.Lock()
	storage[shortURL] = origURL
	lock.Unlock()

	return nil
}

func (s inMemoryStorage) Get(shortURL string) (string, bool, error) {
	lock.RLock()
	v, ok := storage[shortURL]
	lock.RUnlock()

	return v, ok, nil
}
