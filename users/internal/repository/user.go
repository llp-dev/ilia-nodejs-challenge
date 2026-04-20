package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"users/internal/models"
)

var (
	ErrEmailTaken   = errors.New("email already taken")
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{dbPool: dbPool}
}

func (r *UserRepository) Create(ctx context.Context, id, name, email, passwordHash string) (*models.User, error) {
	row := r.dbPool.QueryRow(ctx,
		`INSERT INTO users (id, name, email, password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, now(), now())
		 RETURNING id, name, email, created_at, updated_at`,
		id, name, email, passwordHash,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := r.dbPool.QueryRow(ctx,
		`SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`,
		id,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, string, error) {
	row := r.dbPool.QueryRow(ctx,
		`SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)

	var u models.User
	var passwordHash string
	err := row.Scan(&u.ID, &u.Name, &u.Email, &passwordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}

	return &u, passwordHash, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id, name, email string) (*models.User, error) {
	row := r.dbPool.QueryRow(ctx,
		`UPDATE users SET name = $2, email = $3, updated_at = now()
		 WHERE id = $1
		 RETURNING id, name, email, created_at, updated_at`,
		id, name, email,
	)

	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	cmd, err := r.dbPool.Exec(ctx,
		`UPDATE users SET password = $2, updated_at = now() WHERE id = $1`,
		id, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
