package middleware

import (
	"context"
	"net/http"
	"strings"

	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/response"
	"github.com/4yushraman-jpg/auth-service/internal/token"
)

type userContextKey string

const authenticatedUserKey userContextKey = "authenticated_user"

type AuthenticatedUser struct {
	UserID string
	Email  string
	Role   string
}

func JWTAuth(jwtService *token.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := jwtService.Parse(parts[1])
			if err != nil {
				httpErr := appErrors.ToHTTPError(err)
				response.Error(w, httpErr.Status, httpErr.Message)
				return
			}

			ctx := context.WithValue(r.Context(), authenticatedUserKey, AuthenticatedUser{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserKey).(AuthenticatedUser)
	return user, ok
}
