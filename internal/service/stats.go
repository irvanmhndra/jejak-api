package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/irvanmhndra/jejak-api/internal/repository"
)

type StatsService struct {
	statsRepo repository.StatsRepository
	prRepo    repository.PersonalRecordRepository
}

func NewStatsService(statsRepo repository.StatsRepository, prRepo repository.PersonalRecordRepository) *StatsService {
	return &StatsService{statsRepo, prRepo}
}

type PeriodTotals struct {
	Count          int                     `json:"count"`
	DistanceMeters float64                 `json:"distance_meters"`
	MovingSeconds  int64                   `json:"moving_seconds"`
	ElevationGainM float64                 `json:"elevation_gain_m"`
	ByType         []repository.StatBucket `json:"by_type"`
}

type StatsSummary struct {
	Week    PeriodTotals `json:"week"`
	Month   PeriodTotals `json:"month"`
	Year    PeriodTotals `json:"year"`
	AllTime PeriodTotals `json:"all_time"`
}

func (s *StatsService) Summary(ctx context.Context, userID string) (*StatsSummary, error) {
	now := time.Now()
	weekStart := startOfWeek(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	week, err := s.period(ctx, userID, &weekStart)
	if err != nil {
		return nil, err
	}
	month, err := s.period(ctx, userID, &monthStart)
	if err != nil {
		return nil, err
	}
	year, err := s.period(ctx, userID, &yearStart)
	if err != nil {
		return nil, err
	}
	all, err := s.period(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	return &StatsSummary{Week: week, Month: month, Year: year, AllTime: all}, nil
}

func (s *StatsService) period(ctx context.Context, userID string, since *time.Time) (PeriodTotals, error) {
	buckets, err := s.statsRepo.Totals(ctx, userID, since)
	if err != nil {
		return PeriodTotals{}, err
	}
	pt := PeriodTotals{ByType: buckets}
	for _, b := range buckets {
		pt.Count += b.Count
		pt.DistanceMeters += b.DistanceMeters
		pt.MovingSeconds += b.MovingSeconds
		pt.ElevationGainM += b.ElevationGainM
	}
	return pt, nil
}

func (s *StatsService) Records(ctx context.Context, userID string) ([]model.PersonalRecord, error) {
	return s.prRepo.FindByUserID(ctx, userID)
}

// startOfWeek returns Monday 00:00 of the week containing t.
func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // treat Sunday as last day
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}
