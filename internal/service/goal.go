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

type GoalService struct {
	repo      repository.GoalRepository
	statsRepo repository.StatsRepository
}

func NewGoalService(repo repository.GoalRepository, statsRepo repository.StatsRepository) *GoalService {
	return &GoalService{repo, statsRepo}
}

type CreateGoalRequest struct {
	Period       string  `json:"period" validate:"required,oneof=weekly monthly"`
	ActivityType *string `json:"activity_type" validate:"omitempty,oneof=run ride walk hike other"`
	TargetMeters float64 `json:"target_meters" validate:"required,gt=0"`
}

// GoalWithProgress is a goal plus the distance covered in the current period.
type GoalWithProgress struct {
	*model.Goal
	CurrentMeters float64 `json:"current_meters"`
}

func (s *GoalService) Create(ctx context.Context, userID string, req CreateGoalRequest) (*model.Goal, error) {
	g := &model.Goal{
		ID:           uuid.NewString(),
		UserID:       userID,
		Period:       req.Period,
		ActivityType: req.ActivityType,
		TargetMeters: req.TargetMeters,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GoalService) List(ctx context.Context, userID string) ([]GoalWithProgress, error) {
	goals, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	weekStart := startOfWeek(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	weekBuckets, err := s.statsRepo.Totals(ctx, userID, &weekStart)
	if err != nil {
		return nil, err
	}
	monthBuckets, err := s.statsRepo.Totals(ctx, userID, &monthStart)
	if err != nil {
		return nil, err
	}

	out := make([]GoalWithProgress, 0, len(goals))
	for i := range goals {
		g := goals[i]
		buckets := weekBuckets
		if g.Period == model.PeriodMonthly {
			buckets = monthBuckets
		}
		out = append(out, GoalWithProgress{Goal: &g, CurrentMeters: sumBuckets(buckets, g.ActivityType)})
	}
	return out, nil
}

func (s *GoalService) Delete(ctx context.Context, id, userID string) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFound("target tidak ditemukan")
		}
		return err
	}
	return nil
}

func sumBuckets(buckets []repository.StatBucket, activityType *string) float64 {
	var total float64
	for _, b := range buckets {
		if activityType != nil && b.Type != *activityType {
			continue
		}
		total += b.DistanceMeters
	}
	return total
}
