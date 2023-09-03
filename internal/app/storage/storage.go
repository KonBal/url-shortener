package storage

type Storage interface {
	Set(key uint64, value string) error
	Get(key uint64) (string, bool)
	HasKey(key uint64) bool
}
