package operation

import (
	"context"
	"testing"

	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/require"
)

func TestExpand(t *testing.T) {
	baseURL := "http://base"

	tests := map[string]struct {
		short string

		existingEntries []storage.URLEntry

		want        string
		wantErr     bool
		expectedErr string
	}{
		"correct": {
			short:           "abcd",
			existingEntries: []storage.URLEntry{{ShortURL: "abcd", OriginalURL: "http://orig.link"}},
			want:            "http://orig.link",
		},
		"not_found": {
			short:           "abcd",
			existingEntries: []storage.URLEntry{{ShortURL: "blah", OriginalURL: "blah.blah"}},
			wantErr:         true,
			expectedErr:     notFoundError("original URL not found for shortened abcd").Error(),
		},
		"found_deleted": {
			short:           "abcd",
			existingEntries: []storage.URLEntry{{ShortURL: "abcd", OriginalURL: "http://orig.link", Deleted: true}},
			wantErr:         true,
			expectedErr:     deletedError("url for shortened abcd is already deleted").Error(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			st := storage.NewInMemory()
			st.AddMany(ctx, tt.existingEntries, "")

			s := ShortURLService{BaseURL: baseURL, Storage: st}

			got, err := s.Expand(ctx, tt.short)
			if tt.wantErr {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}
