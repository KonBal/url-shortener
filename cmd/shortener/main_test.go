package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonBal/url-shortener/internal/app/operation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shortener struct {
	shortened string
}

func (s shortener) Shorten(ctx context.Context, url string) (string, error) {
	return s.shortened, nil
}

func TestShortenHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		shortURL    string
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "correct",
			request: "/",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				shortURL:    "http://localhost:8080/abcde",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := operation.ShortenHandle(shortener{shortened: tt.want.shortURL})
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			shortened := string(body)

			assert.Equal(t, tt.want.shortURL, shortened)
		})
	}
}

type expander struct {
	expanded string
}

func (s expander) Expand(ctx context.Context, url string) (string, error) {
	return s.expanded, nil
}

func TestExpandHandler(t *testing.T) {
	type want struct {
		location   string
		statusCode int
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "correct",
			request: "http://localhost:8080/abcde",
			want: want{
				location:   "http://practicum.yandex.ru",
				statusCode: http.StatusTemporaryRedirect,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			h := operation.ExpandHandle(expander{expanded: tt.want.location})
			h(w, request)

			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
