package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appErrors "github.com/4yushraman-jpg/auth-service/internal/errors"
	"github.com/4yushraman-jpg/auth-service/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

const (
	createUserQuery = `
		INSERT INTO users (
			id,
			email,
			password_hash,
			role
		)
		VALUES ($1, $2, $3, $4)
	`
	getUserByEmailQuery = `
		SELECT
			id,
			email,
			password_hash,
			role,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`
	getUserByIDQuery = `
		SELECT
			id,
			email,
			password_hash,
			role,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`
)

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *model.User,
) error {
	_, err := r.db.Exec(
		ctx,
		createUserQuery,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return appErrors.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {
	var user model.User

	err := r.db.QueryRow(
		ctx,
		getUserByEmailQuery,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrInvalidCredentials
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id string,
) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, getUserByIDQuery, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrUnauthorized
		}
		return nil, err
	}
	return &user, nil
}
