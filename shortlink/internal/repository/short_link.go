package repository

import (
	"database/sql"
	"fmt"
	"shortlink/internal/metrics"
)

type ShortLinkRepository struct {
	db *sql.DB
}

func NewShortLinkRepository(db *sql.DB) *ShortLinkRepository {
	return &ShortLinkRepository{db: db}
}
func (r *ShortLinkRepository) Create(code string, longURL string) error {
	_, err := r.db.Exec(
		"INSERT INTO short_links (code, long_url) VALUES (?, ?)",
		code,
		longURL,
	)
	if err != nil {
		return fmt.Errorf("insert short link failed: %w", err)
	}
	return nil
}
func (r *ShortLinkRepository) GetLongURLByCode(code string) (string, error) {
	metrics.DBQueriesTotal.Inc()
	var longURL string
	err := r.db.QueryRow(
		"SELECT long_url FROM short_links WHERE code = ?",
		code,
	).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}
