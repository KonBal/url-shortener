package operation

import "github.com/KonBal/url-shortener/internal/app/storage"

type Rand interface {
	Next() uint64
}
type Encoder interface {
	Encode(v uint64) string
}

type ShortURLService struct {
	BaseURL    string
	Encoder    Encoder
	Storage    storage.Storage
	Uint64Rand Rand
}
