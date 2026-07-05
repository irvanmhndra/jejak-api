package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/irvanmhndra/jejak-api/internal/repository"
	"github.com/irvanmhndra/jejak-api/pkg/apperror"
)

type ActivityService struct {
	repo   repository.ActivityRepository
	prRepo repository.PersonalRecordRepository
}

func NewActivityService(repo repository.ActivityRepository, prRepo repository.PersonalRecordRepository) *ActivityService {
	return &ActivityService{repo, prRepo}
}

type CreateGPSRequest struct {
	Type   string             `json:"type" validate:"required,oneof=run ride walk hike other"`
	Title  string             `json:"title"`
	Notes  *string            `json:"notes"`
	Points []model.TrackPoint `json:"points" validate:"required,min=2"`
}

type CreateManualRequest struct {
	Type           string    `json:"type" validate:"required,oneof=run ride walk hike other"`
	Title          string    `json:"title"`
	Notes          *string   `json:"notes"`
	StartedAt      time.Time `json:"started_at" validate:"required"`
	ElapsedSeconds int       `json:"elapsed_seconds" validate:"required,min=1"`
	DistanceMeters float64   `json:"distance_meters" validate:"min=0"`
}

type UpdateActivityRequest struct {
	Type  string  `json:"type" validate:"required,oneof=run ride walk hike other"`
	Title string  `json:"title" validate:"required"`
	Notes *string `json:"notes"`
}

// ActivityDetail is an activity plus its full GPS track.
type ActivityDetail struct {
	*model.Activity
	Points []model.TrackPoint `json:"points"`
}

func (s *ActivityService) CreateFromGPS(ctx context.Context, userID string, req CreateGPSRequest) (*model.Activity, error) {
	points := req.Points
	stats := computeTrackStats(points)
	a := &model.Activity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Type:           req.Type,
		Title:          defaultTitle(req.Title, req.Type, points[0].T),
		Notes:          req.Notes,
		Source:         model.SourceGPS,
		StartedAt:      points[0].T,
		ElapsedSeconds: stats.ElapsedSeconds,
		MovingSeconds:  stats.MovingSeconds,
		DistanceMeters: stats.DistanceMeters,
		AvgSpeedMps:    stats.AvgSpeedMps,
		MaxSpeedMps:    stats.MaxSpeedMps,
		ElevationGainM: stats.ElevationGainM,
		ElevationLossM: stats.ElevationLossM,
		Polyline:       encodePolyline(points),
		PointCount:     len(points),
	}
	if err := s.repo.Create(ctx, a, &model.ActivityTrack{ActivityID: a.ID, Points: points}); err != nil {
		return nil, err
	}
	s.recomputeRecords(ctx, a)
	return a, nil
}

func (s *ActivityService) CreateManual(ctx context.Context, userID string, req CreateManualRequest) (*model.Activity, error) {
	var avg float64
	if req.ElapsedSeconds > 0 {
		avg = req.DistanceMeters / float64(req.ElapsedSeconds)
	}
	a := &model.Activity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Type:           req.Type,
		Title:          defaultTitle(req.Title, req.Type, req.StartedAt),
		Notes:          req.Notes,
		Source:         model.SourceManual,
		StartedAt:      req.StartedAt,
		ElapsedSeconds: req.ElapsedSeconds,
		MovingSeconds:  req.ElapsedSeconds,
		DistanceMeters: req.DistanceMeters,
		AvgSpeedMps:    avg,
	}
	if err := s.repo.Create(ctx, a, nil); err != nil {
		return nil, err
	}
	s.recomputeRecords(ctx, a)
	return a, nil
}

func (s *ActivityService) ImportGPX(ctx context.Context, userID string, data []byte, typeOverride string) (*model.Activity, error) {
	parsed, err := parseGPX(data)
	if err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	typ := typeOverride
	if typ == "" {
		typ = normalizeGPXType(parsed.Type)
	}
	stats := computeTrackStats(parsed.Points)
	a := &model.Activity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Type:           typ,
		Title:          defaultTitle(parsed.Name, typ, parsed.StartedAt),
		Source:         model.SourceGPX,
		StartedAt:      parsed.StartedAt,
		ElapsedSeconds: stats.ElapsedSeconds,
		MovingSeconds:  stats.MovingSeconds,
		DistanceMeters: stats.DistanceMeters,
		AvgSpeedMps:    stats.AvgSpeedMps,
		MaxSpeedMps:    stats.MaxSpeedMps,
		ElevationGainM: stats.ElevationGainM,
		ElevationLossM: stats.ElevationLossM,
		Polyline:       encodePolyline(parsed.Points),
		PointCount:     len(parsed.Points),
	}
	if err := s.repo.Create(ctx, a, &model.ActivityTrack{ActivityID: a.ID, Points: parsed.Points}); err != nil {
		return nil, err
	}
	s.recomputeRecords(ctx, a)
	return a, nil
}

func (s *ActivityService) List(ctx context.Context, p repository.ActivityListParams) ([]model.Activity, int, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	items, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, p.UserID, p.Type)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *ActivityService) Get(ctx context.Context, id, userID string) (*ActivityDetail, error) {
	a, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, apperror.NotFound("aktivitas tidak ditemukan")
	}
	track, err := s.repo.FindTrack(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return &ActivityDetail{Activity: a, Points: track.Points}, nil
}

func (s *ActivityService) Update(ctx context.Context, id, userID string, req UpdateActivityRequest) (*model.Activity, error) {
	a, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, apperror.NotFound("aktivitas tidak ditemukan")
	}
	a.Type = req.Type
	a.Title = req.Title
	a.Notes = req.Notes
	if err := s.repo.UpdateMeta(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ActivityService) Delete(ctx context.Context, id, userID string) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFound("aktivitas tidak ditemukan")
		}
		return err
	}
	return nil
}

func (s *ActivityService) ExportGPX(ctx context.Context, id, userID string) (string, string, error) {
	a, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return "", "", err
	}
	if a == nil {
		return "", "", apperror.NotFound("aktivitas tidak ditemukan")
	}
	track, err := s.repo.FindTrack(ctx, a.ID)
	if err != nil {
		return "", "", err
	}
	return buildGPX(a, track.Points), a.Title, nil
}

// recomputeRecords compares an activity's summary against the user's PRs for
// that activity type and upserts any it beats. Best-effort (never fails saves).
func (s *ActivityService) recomputeRecords(ctx context.Context, a *model.Activity) {
	type cand struct {
		recordType string
		value      float64
	}
	cands := []cand{
		{model.RecordLongestDistance, a.DistanceMeters},
		{model.RecordLongestDuration, float64(a.MovingSeconds)},
		{model.RecordMostElevation, a.ElevationGainM},
	}
	if a.DistanceMeters >= 400 && a.MovingSeconds >= 60 {
		cands = append(cands, cand{model.RecordFastestSpeed, a.AvgSpeedMps})
	}
	for _, c := range cands {
		if c.value <= 0 {
			continue
		}
		existing, err := s.prRepo.Find(ctx, a.UserID, a.Type, c.recordType)
		if err != nil {
			continue
		}
		if existing != nil && existing.Value >= c.value {
			continue
		}
		_ = s.prRepo.Upsert(ctx, &model.PersonalRecord{
			ID:           uuid.NewString(),
			UserID:       a.UserID,
			ActivityType: a.Type,
			RecordType:   c.recordType,
			Value:        c.value,
			ActivityID:   a.ID,
			AchievedAt:   a.StartedAt,
		})
	}
}

func defaultTitle(title, typ string, at time.Time) string {
	if title != "" {
		return title
	}
	label := map[string]string{
		model.TypeRun:  "Lari",
		model.TypeRide: "Bersepeda",
		model.TypeWalk: "Jalan",
		model.TypeHike: "Hiking",
	}[typ]
	if label == "" {
		label = "Aktivitas"
	}
	var part string
	switch h := at.Hour(); {
	case h < 11:
		part = "Pagi"
	case h < 15:
		part = "Siang"
	case h < 18:
		part = "Sore"
	default:
		part = "Malam"
	}
	return label + " " + part
}

func normalizeGPXType(t string) string {
	switch t {
	case model.TypeRun, "running", "9", "Run":
		return model.TypeRun
	case model.TypeRide, "cycling", "biking", "1", "Ride":
		return model.TypeRide
	case model.TypeWalk, "walking", "Walk":
		return model.TypeWalk
	case model.TypeHike, "hiking", "Hike":
		return model.TypeHike
	default:
		return model.TypeRun
	}
}
