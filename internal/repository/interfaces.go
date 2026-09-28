package repository

import (
	"context"
	"time"

	"github.com/irvanmhndra/jejak-api/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type UserSessionRepository interface {
	Create(ctx context.Context, s *model.UserSession) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.UserSession, error)
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*model.UserSession, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	// Rotate deletes the session oldID and creates next in one transaction.
	// It returns false, creating nothing, when oldID is already gone, so a
	// refresh token can be exchanged only once even under concurrent use.
	Rotate(ctx context.Context, oldID string, next *model.UserSession) (bool, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

type SettingRepository interface {
	FindByUserID(ctx context.Context, userID string) (*model.Setting, error)
	Upsert(ctx context.Context, s *model.Setting) error
}

type ActivityListParams struct {
	UserID string
	Type   string // optional filter
	Limit  int
	Offset int
}

type ActivityRepository interface {
	Create(ctx context.Context, a *model.Activity, track *model.ActivityTrack) error
	FindByID(ctx context.Context, id, userID string) (*model.Activity, error)
	FindTrack(ctx context.Context, activityID string) (*model.ActivityTrack, error)
	List(ctx context.Context, p ActivityListParams) ([]model.Activity, error)
	Count(ctx context.Context, userID, typ string) (int, error)
	UpdateMeta(ctx context.Context, a *model.Activity) error
	Delete(ctx context.Context, id, userID string) error
}

type GoalRepository interface {
	Create(ctx context.Context, g *model.Goal) error
	FindByUserID(ctx context.Context, userID string) ([]model.Goal, error)
	Delete(ctx context.Context, id, userID string) error
}

type PersonalRecordRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]model.PersonalRecord, error)
	Find(ctx context.Context, userID, activityType, recordType string) (*model.PersonalRecord, error)
	Upsert(ctx context.Context, r *model.PersonalRecord) error
	DeleteByActivityID(ctx context.Context, activityID string) error
}

// StatBucket is an aggregate row used by the stats service.
type StatBucket struct {
	Type           string  `db:"type" json:"type"`
	Count          int     `db:"count" json:"count"`
	DistanceMeters float64 `db:"distance_meters" json:"distance_meters"`
	MovingSeconds  int64   `db:"moving_seconds" json:"moving_seconds"`
	ElevationGainM float64 `db:"elevation_gain_m" json:"elevation_gain_m"`
}

type StatsRepository interface {
	// Totals aggregates activities since the given time (nil = all time).
	Totals(ctx context.Context, userID string, since *time.Time) ([]StatBucket, error)
}
