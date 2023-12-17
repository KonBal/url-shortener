package operation

import "github.com/KonBal/url-shortener/internal/app/storage"

// Returns random number.
type Rand interface {
	Next() uint64
}

// Returns encoded value.
type Encoder interface {
	Encode(v uint64) string
}

// Service for managing urls.
type ShortURLService struct {
	BaseURL    string
	Encoder    Encoder
	Storage    storage.Storage
	Uint64Rand Rand
}
