package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type jsonFileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func newFileWriter(fname string) (*jsonFileWriter, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("storage: failed to open the file %s: %w", fname, err)
	}

	return &jsonFileWriter{file: file, encoder: json.NewEncoder(file)}, nil
}

func (w *jsonFileWriter) Close() error {
	return w.file.Close()
}

type fileEntry struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedBy   string `json:"created_by"`
}

func (w *jsonFileWriter) Write(row fileEntry) error {
	return w.encoder.Encode(row)
}

type jsonFileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func newFileReader(fname string) (*jsonFileReader, error) {
	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("storage: failed to open the file %s: %w", fname, err)
	}

	return &jsonFileReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (r *jsonFileReader) Close() error {
	return r.file.Close()
}

func (r *jsonFileReader) Read() (*fileEntry, error) {
	var row fileEntry
	if err := r.decoder.Decode(&row); err != nil {
		return nil, err
	}

	return &row, nil
}

type FileStorage struct {
	fname  string
	writer *jsonFileWriter
	idGen  interface {
		Next() uint64
	}
}

func NewFileStorage(fname string, idGen interface {
	Next() uint64
}) (*FileStorage, error) {
	w, err := newFileWriter(fname)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		fname:  fname,
		writer: w,
		idGen:  idGen,
	}, nil
}

func (s *FileStorage) Close() error {
	return s.writer.Close()
}

func (s *FileStorage) Add(ctx context.Context, u URLEntry, userID string) error {
	return s.writer.Write(fileEntry{
		UUID:        s.idGen.Next(),
		ShortURL:    u.ShortURL,
		OriginalURL: u.OriginalURL,
		CreatedBy:   userID,
	})
}

func (s *FileStorage) AddMany(ctx context.Context, urls []URLEntry, userID string) error {
	for _, u := range urls {
		s.Add(ctx, u, userID)
	}

	return nil
}

func (s *FileStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	entry, err := reader.find(func(entry *fileEntry) bool { return entry.ShortURL == shortURL })
	if err != nil {
		return "", err
	}

	return entry.OriginalURL, nil
}

func (s *FileStorage) GetShort(ctx context.Context, origURL string) (string, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	entry, err := reader.find(func(entry *fileEntry) bool { return entry.OriginalURL == origURL })
	if err != nil {
		return "", err
	}

	return entry.ShortURL, nil
}

func (s *FileStorage) GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var urls []URLEntry
	for {
		entry, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("file: %w", err)
		}

		if entry.CreatedBy == userID {
			urls = append(urls, URLEntry{ShortURL: entry.ShortURL, OriginalURL: entry.OriginalURL})
		}
	}

	return urls, nil
}

func (r *jsonFileReader) find(cond func(entry *fileEntry) bool) (*fileEntry, error) {
	for {
		entry, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, ErrNotFound
			}
			return nil, err
		}

		if cond(entry) {
			return entry, nil
		}
	}
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}
