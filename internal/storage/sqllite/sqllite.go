package sqllite

import (
	"RestApi/internal/storage"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) &&
			errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, err)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM  url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return storage.ErrURLNotFound
	}

	return err
}
