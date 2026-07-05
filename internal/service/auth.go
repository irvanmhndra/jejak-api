package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/irvanmhndra/jejak-api/internal/repository"
	"github.com/irvanmhndra/jejak-api/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.UserSessionRepository
	settingRepo repository.SettingRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.UserSessionRepository,
	settingRepo repository.SettingRepository,
) *AuthService {
	return &AuthService{userRepo, sessionRepo, settingRepo}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.Conflict("email sudah terdaftar")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.NewString(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = s.settingRepo.Upsert(ctx, &model.Setting{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Units:  "metric",
	})

	return s.createSession(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.Unauthorized("email atau password salah")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("email atau password salah")
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, token)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	session, err := s.sessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if session == nil || time.Now().After(session.RefreshExpiresAt) {
		return nil, apperror.TokenExpired()
	}
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, apperror.Unauthorized("user tidak ditemukan")
	}
	_ = s.sessionRepo.DeleteByToken(ctx, session.Token)
	return s.createSession(ctx, user)
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil || time.Now().After(session.ExpiresAt) {
		return nil, apperror.TokenExpired()
	}
	return s.userRepo.FindByID(ctx, session.UserID)
}

func (s *AuthService) Me(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 30 * 24 * time.Hour
)

func (s *AuthService) createSession(ctx context.Context, user *model.User) (*AuthResponse, error) {
	token, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	refresh, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	session := &model.UserSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		Token:            token,
		RefreshToken:     refresh,
		ExpiresAt:        now.Add(accessTokenTTL),
		RefreshExpiresAt: now.Add(refreshTokenTTL),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return &AuthResponse{AccessToken: token, RefreshToken: refresh, User: user}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
