package storage

type inMemoryStorage map[uint64]string

var storage inMemoryStorage

func NewInMemory() Storage {
	storage = make(map[uint64]string)
	return storage
}

func (s inMemoryStorage) Set(key uint64, value string) error {
	storage[key] = value
	return nil
}

func (s inMemoryStorage) Get(key uint64) (string, bool) {
	v, ok := storage[key]
	return v, ok
}

func (s inMemoryStorage) HasKey(key uint64) bool {
	_, ok := storage[key]
	return ok
}
