package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4yushraman-jpg/auth-service/internal/dto"
	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
)

type fakeAuthService struct {
	loginResp *dto.AuthTokensResponse
	loginErr  error
}

func (f *fakeAuthService) Register(ctx context.Context, email, password string) error { return nil }
func (f *fakeAuthService) Login(ctx context.Context, email, password string) (*dto.AuthTokensResponse, error) {
	return f.loginResp, f.loginErr
}
func (f *fakeAuthService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthTokensResponse, error) {
	return nil, nil
}
func (f *fakeAuthService) Logout(ctx context.Context, refreshToken string) error { return nil }

func TestAuthHandlerLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{name: "invalid json", body: "{", wantStatus: http.StatusBadRequest},
		{name: "invalid credentials", body: `{"email":"u@test.com","password":"password123"}`, serviceErr: appErrors.ErrInvalidCredentials, wantStatus: http.StatusUnauthorized},
		{name: "success", body: `{"email":"u@test.com","password":"password123"}`, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewAuthHandler(&fakeAuthService{
				loginResp: &dto.AuthTokensResponse{AccessToken: "a", RefreshToken: "r"},
				loginErr:  tc.serviceErr,
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			h.Login(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
