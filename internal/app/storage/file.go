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

// Close closes file writer.
func (w *jsonFileWriter) Close() error {
	return w.file.Close()
}

type fileEntry struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedBy   string `json:"created_by"`
	Deleted     bool   `json:"deleted,omitempty"`
}

// Wtire writes to json file writer.
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

// Close closes json file reader.
func (r *jsonFileReader) Close() error {
	return r.file.Close()
}

// Read reads from json file reader.
func (r *jsonFileReader) Read() (*fileEntry, error) {
	var row fileEntry
	if err := r.decoder.Decode(&row); err != nil {
		return nil, err
	}

	return &row, nil
}

// File storage.
type FileStorage struct {
	fname  string
	writer *jsonFileWriter
	idGen  interface {
		Next() uint64
	}
}

// NewFileStorage created new file storage.
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

// Close closes file writer.
func (s *FileStorage) Close() error {
	return s.writer.Close()
}

// Add adds new entry to the file.
func (s *FileStorage) Add(ctx context.Context, u URLEntry, userID string) error {
	return s.writer.Write(fileEntry{
		UUID:        s.idGen.Next(),
		ShortURL:    u.ShortURL,
		OriginalURL: u.OriginalURL,
		CreatedBy:   userID,
	})
}

// AddMany adds new entries to the file.
func (s *FileStorage) AddMany(ctx context.Context, urls []URLEntry, userID string) error {
	for _, u := range urls {
		s.Add(ctx, u, userID)
	}

	return nil
}

// GetByShort retrieves a file entry by short url.
func (s *FileStorage) GetByShort(ctx context.Context, shortURL string) (*URLEntry, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	entry, err := reader.find(func(entry *fileEntry) bool { return entry.ShortURL == shortURL })
	if err != nil {
		return nil, err
	}

	return &URLEntry{ShortURL: entry.ShortURL, OriginalURL: entry.OriginalURL, Deleted: entry.Deleted}, nil
}

// GetByOriginal retrieves a file entry by original url.
func (s *FileStorage) GetByOriginal(ctx context.Context, origURL string) (*URLEntry, error) {
	reader, err := newFileReader(s.fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	entry, err := reader.find(func(entry *fileEntry) bool { return entry.OriginalURL == origURL })
	if err != nil {
		return nil, err
	}

	return &URLEntry{ShortURL: entry.ShortURL, OriginalURL: entry.OriginalURL, Deleted: entry.Deleted}, nil
}

// GetURLsCreatedBy retrieves file entries added by user.
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
			urls = append(urls, URLEntry{ShortURL: entry.ShortURL, OriginalURL: entry.OriginalURL, Deleted: entry.Deleted})
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

// MarkDeleted sets deleted flag for given urls.
func (s *FileStorage) MarkDeleted(ctx context.Context, urls ...EntryToDelete) error {
	data, err := os.ReadFile(s.fname)
	if err != nil {
		return fmt.Errorf("file: cannot read file: %w", err)
	}

	var entries []fileEntry

	err = json.Unmarshal(data, &entries)
	if err != nil {
		return fmt.Errorf("file: cannot unmarshal data: %w", err)
	}

	find := func(userID, url string) (*fileEntry, bool) {
		for _, u := range entries {
			if u.CreatedBy == userID && u.ShortURL == url {
				return &u, true
			}
		}

		return nil, false
	}

	for _, u := range urls {
		entry, ok := find(u.UserID, u.ShortURL)
		if ok {
			entry.Deleted = true
		}
	}

	json, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("file: cannot marshal data: %w", err)
	}

	if err := os.WriteFile(s.fname, json, 0666); err != nil {
		return fmt.Errorf("file: cannot write data: %w", err)
	}

	return nil
}

// Ping always returns nil.
func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}
