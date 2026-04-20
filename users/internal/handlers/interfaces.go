package handlers

import (
	"context"

	"users/internal/models"
)

type userRepository interface {
	Create(ctx context.Context, id, name, email, passwordHash string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, string, error)
	UpdateProfile(ctx context.Context, id, name, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
}
