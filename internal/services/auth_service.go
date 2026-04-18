package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rinjold/go-etl-studio/internal/security"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

var ErrBadCredentials = errors.New("invalid email or password")

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (models.User, error)
	Create(ctx context.Context, u models.User) (models.User, error)
}

type AuthService struct {
	userRepo  UserRepository
	jwtSecret string
}

func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, storage.ErrUserNotFound) {
		return "", ErrBadCredentials
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrBadCredentials
	}

	return security.GenerateToken(s.jwtSecret, user.ID, user.Email, user.Role)
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	return s.userRepo.Create(ctx, models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	})
}
