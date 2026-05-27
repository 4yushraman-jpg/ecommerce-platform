package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/model"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

const (
	createRefreshTokenQuery = `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5)
	`
	getRefreshTokenQuery = `
		SELECT id, user_id, token, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token = $1
	`
	revokeRefreshTokenQuery = `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = $1
	`
)

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	_, err := r.db.Exec(
		ctx,
		createRefreshTokenQuery,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.Revoked,
	)
	return err
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.QueryRow(ctx, getRefreshTokenQuery, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.Revoked,
		&rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrTokenNotFound
		}
		return nil, err
	}
	if rt.Revoked {
		return nil, appErrors.ErrTokenRevoked
	}
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, appErrors.ErrTokenExpired
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	cmd, err := r.db.Exec(ctx, revokeRefreshTokenQuery, token)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return appErrors.ErrTokenNotFound
	}
	return nil
}
