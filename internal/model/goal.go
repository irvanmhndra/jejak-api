package model

import "time"

// Goal periods.
const (
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
)

// Goal is a per-period distance target, optionally scoped to an activity type.
type Goal struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"-"`
	Period       string    `db:"period" json:"period"`
	ActivityType *string   `db:"activity_type" json:"activity_type,omitempty"`
	TargetMeters float64   `db:"target_meters" json:"target_meters"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
