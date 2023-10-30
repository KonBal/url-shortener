package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (s *DBStorage) Add(ctx context.Context, u URLEntry, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`insert into urls(short_url, original_url, created_by) values ($1, $2, $3)`,
		u.ShortURL, u.OriginalURL, userID)

	if err != nil {
		var pgErr *pgconn.PgError
		c := errors.As(err, &pgErr)
		if c && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrNotUnique
		}

		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *DBStorage) AddMany(ctx context.Context, urls []URLEntry, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `insert into urls(short_url, original_url, created_by) values ($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	for _, u := range urls {
		_, err := stmt.ExecContext(ctx, u.ShortURL, u.OriginalURL, userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db: %w", err)
		}
	}

	return tx.Commit()
}

func (s *DBStorage) GetByShort(ctx context.Context, shortURL string) (*URLEntry, error) {
	const query = `
		select u.original_url, u.short_url, u.deleted
		from urls as u
		where u.short_url = $1;
	`

	var u URLEntry

	err := s.db.QueryRowContext(ctx, query, shortURL).Scan(&u.OriginalURL, &u.ShortURL, &u.Deleted)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("db: %w", err)
	}

	return &u, nil
}

func (s *DBStorage) GetByOriginal(ctx context.Context, origURL string) (*URLEntry, error) {
	const query = `
		select u.original_url, u.short_url, u.deleted
		from urls as u
		where u.original_url = $1;
	`

	var u URLEntry

	err := s.db.QueryRowContext(ctx, query, origURL).Scan(&u.OriginalURL, &u.ShortURL, &u.Deleted)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return &u, nil
}

func (s *DBStorage) GetURLsCreatedBy(ctx context.Context, userID string) ([]URLEntry, error) {
	const query = `
		select u.short_url, u.original_url, u.deleted
		from urls as u
		where u.created_by = $1;
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	var urls []URLEntry

	for rows.Next() && err == nil {
		var u URLEntry
		err = rows.Scan(&u.ShortURL, &u.OriginalURL, &u.Deleted)
		urls = append(urls, u)
	}

	if err == nil {
		err = rows.Err()
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return []URLEntry{}, nil
	case err != nil:
		return nil, fmt.Errorf("db: %w", err)
	}

	return urls, nil
}

func (s *DBStorage) MarkDeleted(ctx context.Context, urls ...EntryToDelete) error {
	var conditions []string
	var args []any

	for i, u := range urls {
		base := i * 2
		c := fmt.Sprintf("(u.created_by = $%d and u.short_url = $%d)", base+1, base+2)
		conditions = append(conditions, c)
		args = append(args, u.UserID, u.ShortURL)
	}

	query := `
		update urls as u
		set deleted = true
		where ` + strings.Join(conditions, " or ") + ";"

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *DBStorage) Bootstrap(migrationFiles fs.FS) error {
	if err := applyMigrations(s.db, migrationFiles); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func applyMigrations(db *sql.DB, fsys fs.FS) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set up migration tool: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
