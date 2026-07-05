package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/irvanmhndra/jejak-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type activityRepo struct{ db *sqlx.DB }

func NewActivityRepository(db *sqlx.DB) *activityRepo { return &activityRepo{db} }

func (r *activityRepo) Create(ctx context.Context, a *model.Activity, track *model.ActivityTrack) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO activities (
			id, user_id, type, title, notes, source, started_at,
			elapsed_seconds, moving_seconds, distance_meters,
			avg_speed_mps, max_speed_mps, elevation_gain_m, elevation_loss_m,
			calories, polyline, point_count
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		a.ID, a.UserID, a.Type, a.Title, a.Notes, a.Source, a.StartedAt,
		a.ElapsedSeconds, a.MovingSeconds, a.DistanceMeters,
		a.AvgSpeedMps, a.MaxSpeedMps, a.ElevationGainM, a.ElevationLossM,
		a.Calories, a.Polyline, a.PointCount,
	)
	if err != nil {
		return err
	}

	if track != nil && len(track.Points) > 0 {
		pts, err := json.Marshal(track.Points)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx,
			`INSERT INTO activity_tracks (activity_id, points, point_count) VALUES ($1, $2, $3)`,
			a.ID, pts, len(track.Points),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *activityRepo) FindByID(ctx context.Context, id, userID string) (*model.Activity, error) {
	var a model.Activity
	err := r.db.GetContext(ctx, &a,
		`SELECT * FROM activities WHERE id = $1 AND user_id = $2`, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &a, err
}

func (r *activityRepo) FindTrack(ctx context.Context, activityID string) (*model.ActivityTrack, error) {
	var raw []byte
	err := r.db.GetContext(ctx, &raw,
		`SELECT points FROM activity_tracks WHERE activity_id = $1`, activityID)
	if errors.Is(err, sql.ErrNoRows) {
		return &model.ActivityTrack{ActivityID: activityID, Points: []model.TrackPoint{}}, nil
	}
	if err != nil {
		return nil, err
	}
	track := &model.ActivityTrack{ActivityID: activityID, Points: []model.TrackPoint{}}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &track.Points); err != nil {
			return nil, err
		}
	}
	return track, nil
}

func (r *activityRepo) List(ctx context.Context, p repository.ActivityListParams) ([]model.Activity, error) {
	list := []model.Activity{}
	if p.Type != "" {
		err := r.db.SelectContext(ctx, &list,
			`SELECT * FROM activities WHERE user_id = $1 AND type = $2
			 ORDER BY started_at DESC LIMIT $3 OFFSET $4`,
			p.UserID, p.Type, p.Limit, p.Offset)
		return list, err
	}
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM activities WHERE user_id = $1
		 ORDER BY started_at DESC LIMIT $2 OFFSET $3`,
		p.UserID, p.Limit, p.Offset)
	return list, err
}

func (r *activityRepo) Count(ctx context.Context, userID, typ string) (int, error) {
	var n int
	if typ != "" {
		err := r.db.GetContext(ctx, &n,
			`SELECT COUNT(*) FROM activities WHERE user_id = $1 AND type = $2`, userID, typ)
		return n, err
	}
	err := r.db.GetContext(ctx, &n, `SELECT COUNT(*) FROM activities WHERE user_id = $1`, userID)
	return n, err
}

func (r *activityRepo) UpdateMeta(ctx context.Context, a *model.Activity) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE activities SET type = $1, title = $2, notes = $3, updated_at = NOW()
		 WHERE id = $4 AND user_id = $5`,
		a.Type, a.Title, a.Notes, a.ID, a.UserID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *activityRepo) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM activities WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
