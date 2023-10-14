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

func NewDBStorage(connString string) (*DBStorage, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("db: failed to establish connection: %w", err)
	} else if err = db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("db: failed to create tables: %w", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func (s *DBStorage) Add(ctx context.Context, uuid uint64, shortURL string, origURL string) error {
	_, err := s.db.ExecContext(ctx,
		`insert into urls(uuid, short_url, original_url) values ($1, $2, $3)`, uuid, shortURL, origURL)
	if err != nil {
		return fmt.Errorf("db: failed to add entry: %w", err)
	}

	return nil
}

func (s *DBStorage) Get(ctx context.Context, shortURL string) (string, bool, error) {
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
		return "", false, fmt.Errorf("db: %v", err)
	}

	return url, true, nil
}

const createTablesQuery = `
create table if not exists urls (
	id serial primary key,
	uuid numeric not null unique,
	short_url varchar not null,
	original_url varchar not null
);`

func createTables(db *sql.DB) error {
	_, err := db.Exec(createTablesQuery)
	return err
}
