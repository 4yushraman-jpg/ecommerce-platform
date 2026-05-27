package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/4yushraman-jpg/auth-service/internal/dto"
	"github.com/4yushraman-jpg/auth-service/internal/model"
	"github.com/4yushraman-jpg/auth-service/internal/token"
)

type AuthService struct {
	userRepo         userRepository
	refreshTokenRepo refreshTokenRepository
	passwordService  passwordService
	tokenService     *token.Service
	refreshTTL       time.Duration
}

type userRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type refreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
}

type passwordService interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

func NewAuthService(
	userRepo userRepository,
	refreshTokenRepo refreshTokenRepository,
	passwordService passwordService,
	tokenService *token.Service,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordService:  passwordService,
		tokenService:     tokenService,
		refreshTTL:       refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	hashedPassword, err := s.passwordService.Hash(password)
	if err != nil {
		return err
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}
	return s.userRepo.Create(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*dto.AuthTokensResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := s.passwordService.Compare(user.PasswordHash, password); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthTokensResponse, error) {
	if _, err := s.tokenService.Parse(refreshToken); err != nil {
		return nil, err
	}

	dbToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, dbToken.UserID.String())
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshTokenRepo.RevokeByToken(ctx, refreshToken)
}

func (s *AuthService) issueTokens(ctx context.Context, user *model.User) (*dto.AuthTokensResponse, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokenRepo.Create(ctx, &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(s.refreshTTL),
		Revoked:   false,
	}); err != nil {
		return nil, err
	}
	return &dto.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
