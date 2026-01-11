package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/nanda/doit/modules/core/entity"
)

type PostgresURLAnalyticRepo struct {
	db *sql.DB
}

func NewPostgresURLAnalyticRepo(db *sql.DB) *PostgresURLAnalyticRepo {
	return &PostgresURLAnalyticRepo{db: db}
}

func (r *PostgresURLAnalyticRepo) Create(ctx context.Context, analytic *entity.URLAnalytic) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO url_analytics (url_id, long_url, created_at, expires_at, click_count, last_accessed_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		analytic.URLID,
		analytic.LongURL,
		analytic.CreatedAt,
		analytic.ExpiresAt,
		analytic.ClickCount,
		analytic.LastAccessedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PostgresURLAnalyticRepo) GetByURLID(ctx context.Context, urlID int64) (*entity.URLAnalytic, error) {
	var analytic entity.URLAnalytic
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, url_id, long_url, created_at, expires_at, click_count, last_accessed_at
		 FROM url_analytics WHERE url_id = $1`,
		urlID,
	).Scan(
		&analytic.ID,
		&analytic.URLID,
		&analytic.LongURL,
		&analytic.CreatedAt,
		&analytic.ExpiresAt,
		&analytic.ClickCount,
		&analytic.LastAccessedAt,
	)
	if err != nil {
		return nil, err
	}
	return &analytic, nil
}

func (r *PostgresURLAnalyticRepo) UpdateStat(ctx context.Context, urlID int64, now time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE url_analytics SET click_count = click_count + 1, last_accessed_at = $1 WHERE url_id = $2`,
		now,
		urlID,
	)
	return err
}
