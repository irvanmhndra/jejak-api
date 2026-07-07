package postgres

import (
	"context"
	"database/sql"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type goalRepo struct{ db *sqlx.DB }

func NewGoalRepository(db *sqlx.DB) *goalRepo { return &goalRepo{db} }

func (r *goalRepo) Create(ctx context.Context, g *model.Goal) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO goals (id, user_id, period, activity_type, target_meters)
		 VALUES ($1,$2,$3,$4,$5)`,
		g.ID, g.UserID, g.Period, g.ActivityType, g.TargetMeters,
	)
	return err
}

func (r *goalRepo) FindByUserID(ctx context.Context, userID string) ([]model.Goal, error) {
	list := []model.Goal{}
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM goals WHERE user_id = $1 ORDER BY period, created_at`, userID)
	return list, err
}

func (r *goalRepo) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM goals WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
