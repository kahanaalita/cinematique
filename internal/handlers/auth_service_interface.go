package handlers

import "cinematique/internal/auth"

// AuthService Определяет интерфейс для операций аутентификации
type AuthService interface {
	// Register Создает нового пользователя с данными учетными данными
	Register(username, email, password, role string) (int, error)
	// Login Аутентифицирует пользователя и возвращает пару токенов JWT
	Login(username, password string) (*auth.TokenPair, error)
	// RefreshToken обновляет access token с помощью refresh token
	RefreshToken(refreshToken string) (*auth.TokenPair, error)
	// Logout выполняет выход пользователя из системы
	Logout(refreshToken string) error
}
