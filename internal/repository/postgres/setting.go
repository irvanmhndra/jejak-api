package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type settingRepo struct{ db *sqlx.DB }

func NewSettingRepository(db *sqlx.DB) *settingRepo { return &settingRepo{db} }

func (r *settingRepo) FindByUserID(ctx context.Context, userID string) (*model.Setting, error) {
	var s model.Setting
	err := r.db.GetContext(ctx, &s, `SELECT * FROM settings WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *settingRepo) Upsert(ctx context.Context, s *model.Setting) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO settings (id, user_id, units)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET units = EXCLUDED.units, updated_at = NOW()`,
		s.ID, s.UserID, s.Units,
	)
	return err
}
