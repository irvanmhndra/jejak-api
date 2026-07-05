package model

import "time"

// Activity types.
const (
	TypeRun   = "run"
	TypeRide  = "ride"
	TypeWalk  = "walk"
	TypeHike  = "hike"
	TypeOther = "other"
)

// Activity sources.
const (
	SourceGPS    = "gps"
	SourceManual = "manual"
	SourceGPX    = "gpx"
)

// Activity is the denormalized summary of a recorded workout.
type Activity struct {
	ID             string    `db:"id" json:"id"`
	UserID         string    `db:"user_id" json:"-"`
	Type           string    `db:"type" json:"type"`
	Title          string    `db:"title" json:"title"`
	Notes          *string   `db:"notes" json:"notes,omitempty"`
	Source         string    `db:"source" json:"source"`
	StartedAt      time.Time `db:"started_at" json:"started_at"`
	ElapsedSeconds int       `db:"elapsed_seconds" json:"elapsed_seconds"`
	MovingSeconds  int       `db:"moving_seconds" json:"moving_seconds"`
	DistanceMeters float64   `db:"distance_meters" json:"distance_meters"`
	AvgSpeedMps    float64   `db:"avg_speed_mps" json:"avg_speed_mps"`
	MaxSpeedMps    float64   `db:"max_speed_mps" json:"max_speed_mps"`
	ElevationGainM float64   `db:"elevation_gain_m" json:"elevation_gain_m"`
	ElevationLossM float64   `db:"elevation_loss_m" json:"elevation_loss_m"`
	Calories       *int      `db:"calories" json:"calories,omitempty"`
	Polyline       string    `db:"polyline" json:"polyline"`
	PointCount     int       `db:"point_count" json:"point_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// TrackPoint is a single GPS sample. Elevation is optional; Dist is the
// cumulative distance (meters) from the start, filled in server-side.
type TrackPoint struct {
	Lat  float64   `json:"lat"`
	Lng  float64   `json:"lng"`
	Ele  *float64  `json:"ele,omitempty"`
	T    time.Time `json:"t"`
	Dist float64   `json:"dist"`
}

// ActivityTrack holds the full sample stream for one activity (stored as JSONB).
type ActivityTrack struct {
	ActivityID string       `json:"activity_id"`
	Points     []TrackPoint `json:"points"`
}

// Personal record types.
const (
	RecordLongestDistance = "longest_distance" // meters
	RecordLongestDuration = "longest_duration" // moving seconds
	RecordMostElevation   = "most_elevation"   // meters
	RecordFastestSpeed    = "fastest_speed"    // m/s (avg moving)
)

type PersonalRecord struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"-"`
	ActivityType string    `db:"activity_type" json:"activity_type"`
	RecordType   string    `db:"record_type" json:"record_type"`
	Value        float64   `db:"value" json:"value"`
	ActivityID   string    `db:"activity_id" json:"activity_id"`
	AchievedAt   time.Time `db:"achieved_at" json:"achieved_at"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
