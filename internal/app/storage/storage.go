package storage

type Storage interface {
	Add(uuid uint64, shortURL string, origURL string) error
	Get(shortURL string) (string, bool, error)
}
