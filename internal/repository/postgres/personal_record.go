package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type personalRecordRepo struct{ db *sqlx.DB }

func NewPersonalRecordRepository(db *sqlx.DB) *personalRecordRepo { return &personalRecordRepo{db} }

func (r *personalRecordRepo) FindByUserID(ctx context.Context, userID string) ([]model.PersonalRecord, error) {
	list := []model.PersonalRecord{}
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM personal_records WHERE user_id = $1
		 ORDER BY activity_type, record_type`, userID)
	return list, err
}

func (r *personalRecordRepo) Find(ctx context.Context, userID, activityType, recordType string) (*model.PersonalRecord, error) {
	var rec model.PersonalRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM personal_records
		 WHERE user_id = $1 AND activity_type = $2 AND record_type = $3`,
		userID, activityType, recordType)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *personalRecordRepo) Upsert(ctx context.Context, rec *model.PersonalRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO personal_records (id, user_id, activity_type, record_type, value, activity_id, achieved_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (user_id, activity_type, record_type)
		 DO UPDATE SET value = EXCLUDED.value, activity_id = EXCLUDED.activity_id,
		               achieved_at = EXCLUDED.achieved_at, updated_at = NOW()`,
		rec.ID, rec.UserID, rec.ActivityType, rec.RecordType, rec.Value, rec.ActivityID, rec.AchievedAt,
	)
	return err
}

func (r *personalRecordRepo) DeleteByActivityID(ctx context.Context, activityID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM personal_records WHERE activity_id = $1`, activityID)
	return err
}
