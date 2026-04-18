package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, u models.User) (models.User, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, u.ID, u.Email, u.PasswordHash, u.Role).Scan(&u.CreatedAt)
	return u, err
}
