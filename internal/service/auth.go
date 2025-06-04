package service

import (
	"cinematigue/internal/domain"
	"cinematigue/internal/repository"
	"cinematigue/internal/auth"
	"golang.org/x/crypto/bcrypt"
	"errors"
)

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register регистрирует пользователя
func (s *AuthService) Register(username, password, role string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	if role == "" {
		role = domain.RoleUser
	}
	user := domain.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}
	return s.repo.CreateUser(user)
}

// Login аутентифицирует пользователя и возвращает JWT
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return auth.GenerateJWT(user.ID, user.Username, user.Role)
}
