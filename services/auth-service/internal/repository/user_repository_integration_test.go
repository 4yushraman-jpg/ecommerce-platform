package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/4yushraman-jpg/auth-service/internal/model"
)

func TestUserRepositoryIntegration_CreateAndGetByEmail(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	email := "integration-" + uuid.NewString() + "@example.com"
	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "hash",
		Role:         "user",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.Email != email {
		t.Fatalf("expected email %s, got %s", email, got.Email)
	}
}
