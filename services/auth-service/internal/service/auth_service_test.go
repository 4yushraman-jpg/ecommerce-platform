package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/model"
	"github.com/4yushraman-jpg/auth-service/internal/token"
)

type fakeUserRepo struct {
	user *model.User
	err  error
}

func (f *fakeUserRepo) Create(ctx context.Context, user *model.User) error { return nil }
func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return f.user, f.err
}
func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	return f.user, f.err
}

type fakeRefreshRepo struct{ created bool }

func (f *fakeRefreshRepo) Create(ctx context.Context, token *model.RefreshToken) error { f.created = true; return nil }
func (f *fakeRefreshRepo) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	return nil, nil
}
func (f *fakeRefreshRepo) RevokeByToken(ctx context.Context, token string) error { return nil }

type fakePasswordSvc struct{ compareErr error }

func (f *fakePasswordSvc) Hash(password string) (string, error) { return "hash", nil }
func (f *fakePasswordSvc) Compare(hash string, password string) error { return f.compareErr }

func TestAuthServiceLogin(t *testing.T) {
	userID := uuid.New()
	jwtSvc := token.NewJWTService("secret", 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name      string
		repoErr   error
		compareErr error
		wantErr   error
	}{
		{name: "invalid credentials from repo", repoErr: appErrors.ErrInvalidCredentials, wantErr: appErrors.ErrInvalidCredentials},
		{name: "password mismatch", compareErr: appErrors.ErrInvalidCredentials, wantErr: appErrors.ErrInvalidCredentials},
		{name: "success"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewAuthService(
				&fakeUserRepo{
					user: &model.User{ID: userID, Email: "test@example.com", Role: "user", PasswordHash: "hash"},
					err:  tc.repoErr,
				},
				&fakeRefreshRepo{},
				&fakePasswordSvc{compareErr: tc.compareErr},
				jwtSvc,
				7*24*time.Hour,
			)
			got, err := svc.Login(context.Background(), "test@example.com", "password123")
			if tc.wantErr != nil {
				if err == nil || err.Error() != tc.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatalf("expected non-empty tokens")
			}
		})
	}
}
