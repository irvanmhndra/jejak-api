package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/irvanmhndra/jejak-api/internal/repository"
)

type SettingService struct {
	repo repository.SettingRepository
}

func NewSettingService(repo repository.SettingRepository) *SettingService {
	return &SettingService{repo}
}

func (s *SettingService) Get(ctx context.Context, userID string) (*model.Setting, error) {
	setting, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.Setting{ID: uuid.NewString(), UserID: userID, Units: "metric"}
		_ = s.repo.Upsert(ctx, setting)
	}
	return setting, nil
}

type UpdateSettingRequest struct {
	Units string `json:"units" validate:"required,oneof=metric imperial"`
}

func (s *SettingService) Update(ctx context.Context, userID string, req UpdateSettingRequest) (*model.Setting, error) {
	setting := &model.Setting{ID: uuid.NewString(), UserID: userID, Units: req.Units}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return s.repo.FindByUserID(ctx, userID)
}
