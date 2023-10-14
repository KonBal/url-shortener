package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type URLEntry struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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

func (w *jsonFileWriter) Write(row URLEntry) error {
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

func (r *jsonFileReader) Read() (*URLEntry, error) {
	var row URLEntry
	if err := r.decoder.Decode(&row); err != nil {
		return nil, err
	}

	return &row, nil
}

type FileStorage struct {
	fname  string
	writer *jsonFileWriter
}

func NewFileStorage(fname string) (*FileStorage, error) {
	w, err := newFileWriter(fname)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		fname:  fname,
		writer: w,
	}, nil
}

func (s *FileStorage) Close() error {
	return s.writer.Close()
}

func (s *FileStorage) Add(ctx context.Context, uuid uint64, shortURL string, origURL string) error {
	return s.writer.Write(URLEntry{UUID: uuid, ShortURL: shortURL, OriginalURL: origURL})
}

func (s *FileStorage) Get(ctx context.Context, shortURL string) (string, bool, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return "", false, err
	}
	defer reader.Close()

	for {
		entry, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", false, nil
			}
			return "", false, err
		}

		if entry.ShortURL == shortURL {
			return entry.OriginalURL, true, nil
		}
	}
}
