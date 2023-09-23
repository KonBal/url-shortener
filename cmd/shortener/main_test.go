package main

import (
	"bytes"
	"context"
	"encoding/json"
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
		name         string
		request      string
		body         string
		shortURLHost string
		shortener    shortener
		want         want
	}{
		{
			name:         "correct",
			request:      "/",
			body:         "abcde",
			shortURLHost: "localhost:8080",
			shortener:    shortener{shortened: "abcde"},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				shortURL:    "http://localhost:8080/abcde",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			h := operation.ShortenHandle(tt.shortURLHost, tt.shortener)
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

func TestShortenFromJSONHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		resp        struct{ Result string }
	}

	tests := []struct {
		name         string
		request      string
		body         struct{ URL string }
		shortURLHost string
		shortener    shortener
		want         want
	}{
		{
			name:         "correct",
			request:      "/shorten",
			body:         struct{ URL string }{URL: "abcde"},
			shortURLHost: "localhost:8080",
			shortener:    shortener{shortened: "abcde"},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
				resp:        struct{ Result string }{Result: "http://localhost:8080/abcde"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.body)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, tt.request, &body)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := operation.ShortenFromJSONHandle(tt.shortURLHost, tt.shortener)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			var resp struct {
				Result string
			}

			err = json.NewDecoder(result.Body).Decode(&resp)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.resp, resp)
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
		name     string
		request  string
		expander expander
		want     want
	}{
		{
			name:     "correct",
			request:  "http://localhost:8080/abcde",
			expander: expander{expanded: "http://practicum.yandex.ru"},
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

			h := operation.ExpandHandle(tt.expander)
			h(w, request)

			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
