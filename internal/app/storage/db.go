package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (s *DBStorage) Add(ctx context.Context, u URLEntry) error {
	_, err := s.db.ExecContext(ctx,
		`insert into urls(short_url, original_url) values ($1, $2)`,
		u.ShortURL, u.OriginalURL)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *DBStorage) AddMany(ctx context.Context, urls []URLEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `insert into urls(short_url, original_url) values ($1, $2)`)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	for _, u := range urls {
		_, err := stmt.ExecContext(ctx, u.ShortURL, u.OriginalURL)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db: %w", err)
		}
	}

	return tx.Commit()
}

func (s *DBStorage) GetOriginal(ctx context.Context, shortURL string) (string, bool, error) {
	const query = `
		select u.original_url
		from urls as u
		where u.short_url = $1;
	`

	var url string

	err := s.db.QueryRowContext(ctx, query, shortURL).Scan(&url)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("db: %w", err)
	}

	return url, true, nil
}

func (s *DBStorage) Bootstrap() error {
	_, err := s.db.Exec(`
		create table if not exists urls (
			id serial primary key,
			short_url varchar not null unique,
			original_url varchar not null
		);
	`)
	if err != nil {
		return fmt.Errorf("db: failed to bootstrap: %w", err)
	}

	return nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
