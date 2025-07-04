package service

import (
	"cinematique/internal/domain"
	"cinematique/internal/auth"
	"cinematique/internal/repository"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register регистрирует пользователя
func (s *AuthService) Register(username, email, password, role string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	if role == "" {
		role = domain.RoleUser
	}
	user := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
	return s.repo.CreateUser(user)
}

// Login проверяет учетные данные и возвращает JWT токены
func (s *AuthService) Login(username, password string) (*auth.TokenPair, error) {
	// Получаем пользователя по имени пользователя
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Генерируем JWT токены
	tokenPair, err := auth.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenPair, nil
}

// RefreshToken обновляет access token с помощью refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	// Валидируем refresh token и получаем claims
	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Получаем пользователя по ID из токена
	user, err := s.repo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Генерируем новую пару токенов
	newTokenPair, err := auth.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %v", err)
	}

	return newTokenPair, nil
}

// Logout выполняет выход пользователя (в текущей реализации просто валидирует токен)
func (s *AuthService) Logout(refreshToken string) error {
	// Валидируем refresh token
	_, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token")
	}

	// В реальном приложении здесь можно добавить логику для добавления токена в черный список
	// или обновления статуса пользователя, если это необходимо

	return nil
}
