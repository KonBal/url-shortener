package operation

import (
	"context"
	"strconv"
	"testing"

	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/require"
)

type prand []uint64

func (r *prand) Next() uint64 {
	var i uint64

	if len(*r) > 0 {
		i, *r = (*r)[0], (*r)[1:]
	}

	return i
}

type encoder struct{}

func (en encoder) Encode(val uint64) string {
	return strconv.Itoa(int(val))
}

func TestShorten(t *testing.T) {
	tests := map[string]struct {
		orig string
		base string

		rand            Rand
		encoder         Encoder
		existingEntries []storage.URLEntry

		want        string
		wantErr     bool
		expectedErr string
	}{
		"empty": {
			orig: "",
			base: "http://base",

			rand:    &prand{0},
			encoder: encoder{},

			want: "http://base" + "/0",
		},
		"not_unique": {
			orig: "link.ru",
			base: "http://base",

			rand:            &prand{12345},
			encoder:         encoder{},
			existingEntries: []storage.URLEntry{{OriginalURL: "link.ru"}},

			wantErr:     true,
			expectedErr: notUniqueError{ShortURL: "http://base" + "/12345"}.Error(),
		},
		"correct": {
			orig: "link.ru",
			base: "http://base",

			rand:    &prand{12345},
			encoder: encoder{},

			want: "http://base" + "/12345",
		},
		"correct_no_schema": {
			orig: "link.ru",
			base: "base",

			rand:    &prand{12345},
			encoder: encoder{},

			want: "http://base" + "/12345",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			st := storage.NewInMemory()
			st.AddMany(ctx, tt.existingEntries, "")

			s := ShortURLService{BaseURL: tt.base, Encoder: tt.encoder, Storage: st, Uint64Rand: tt.rand}

			got, err := s.Shorten(ctx, "", tt.orig)
			if tt.wantErr {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestShortenMany(t *testing.T) {
	tests := map[string]struct {
		orig []CorrelatedOrigURL
		base string

		rand            Rand
		encoder         Encoder
		existingEntries []storage.URLEntry

		want    []CorrelatedShortURL
		wantErr bool
	}{
		"empty": {
			want: []CorrelatedShortURL{},
		},
		"correct": {
			orig:    []CorrelatedOrigURL{{CorrelationID: "1", OrigURL: "http://ab.cd"}, {CorrelationID: "2", OrigURL: "http://ef.gh"}},
			base:    "http://base",
			rand:    &prand{0, 1},
			encoder: encoder{},
			want:    []CorrelatedShortURL{{CorrelationID: "1", ShortURL: "http://base" + "/0"}, {CorrelationID: "2", ShortURL: "http://base" + "/1"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			st := storage.NewInMemory()
			st.AddMany(ctx, tt.existingEntries, "")

			s := ShortURLService{BaseURL: tt.base, Encoder: tt.encoder, Storage: st, Uint64Rand: tt.rand}

			got, err := s.ShortenMany(ctx, "", tt.orig)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
