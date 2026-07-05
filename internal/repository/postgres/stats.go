package postgres

import (
	"context"
	"time"

	"github.com/irvanmhndra/jejak-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type statsRepo struct{ db *sqlx.DB }

func NewStatsRepository(db *sqlx.DB) *statsRepo { return &statsRepo{db} }

func (r *statsRepo) Totals(ctx context.Context, userID string, since *time.Time) ([]repository.StatBucket, error) {
	buckets := []repository.StatBucket{}
	const sel = `SELECT type,
			COUNT(*) AS count,
			COALESCE(SUM(distance_meters), 0) AS distance_meters,
			COALESCE(SUM(moving_seconds), 0) AS moving_seconds,
			COALESCE(SUM(elevation_gain_m), 0) AS elevation_gain_m
		FROM activities WHERE user_id = $1`
	if since != nil {
		err := r.db.SelectContext(ctx, &buckets,
			sel+` AND started_at >= $2 GROUP BY type`, userID, *since)
		return buckets, err
	}
	err := r.db.SelectContext(ctx, &buckets, sel+` GROUP BY type`, userID)
	return buckets, err
}
