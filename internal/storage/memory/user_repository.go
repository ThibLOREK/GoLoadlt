package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type UserRepository struct {
	mu    sync.RWMutex
	store map[string]models.User // keyed by email
}

func NewUserRepository() *UserRepository {
	return &UserRepository{store: make(map[string]models.User)}
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.store[email]
	if !ok {
		return models.User{}, storage.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) Create(_ context.Context, u models.User) (models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u.CreatedAt = time.Now().UTC()
	r.store[u.Email] = u
	return u, nil
}
