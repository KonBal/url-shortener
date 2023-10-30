package operation

import "github.com/KonBal/url-shortener/internal/app/storage"

type ShortURLService struct {
	BaseURL string
	Encoder interface {
		Encode(v uint64) string
	}
	Storage    storage.Storage
	Uint64Rand interface {
		Next() uint64
	}
}
