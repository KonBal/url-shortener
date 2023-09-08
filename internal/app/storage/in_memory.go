package storage

import "sync"

type inMemoryStorage map[uint64]string

var storage inMemoryStorage
var lock *sync.RWMutex

func NewInMemory() Storage {
	storage = make(map[uint64]string)
	lock = &sync.RWMutex{}
	return storage
}

func (s inMemoryStorage) Set(key uint64, value string) error {
	lock.Lock()
	storage[key] = value
	lock.Unlock()

	return nil
}

func (s inMemoryStorage) Get(key uint64) (string, bool) {
	lock.RLock()
	v, ok := storage[key]
	lock.RUnlock()

	return v, ok
}

func (s inMemoryStorage) HasKey(key uint64) bool {
	lock.RLock()
	_, ok := storage[key]
	lock.RUnlock()

	return ok
}
