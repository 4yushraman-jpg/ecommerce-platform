package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4yushraman-jpg/auth-service/internal/token"
)

func TestJWTAuthMiddleware(t *testing.T) {
	jwtSvc := token.NewJWTService("secret", 15*time.Minute, 7*24*time.Hour)
	valid, _ := jwtSvc.GenerateAccessToken(uuid.New(), "user@test.com", "user")

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "missing token", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer invalid", wantStatus: http.StatusUnauthorized},
		{name: "valid token", authHeader: "Bearer " + valid, wantStatus: http.StatusOK},
	}

	handler := JWTAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := GetUserFromContext(r.Context()); !ok {
			t.Fatalf("expected user in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
