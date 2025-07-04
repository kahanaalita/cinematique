package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Конфигурация JWT
const (
	AccessTokenExpiry  = 15 * time.Minute  // Время жизни токена доступа составляет 15 минут
	RefreshTokenExpiry = 7 * 24 * time.Hour // Время жизни токена обновления составляет 7 дней
)

// JWTKey содержит секретный ключ для подписи JWT-токенов
var JWTKey []byte

// InitJWTKey инициализирует ключ JWT из переменной окружения или генерирует новый ключ
func InitJWTKey() error {
	// Попытка получить ключ из переменной окружения
	keyStr := os.Getenv("JWT_SECRET_KEY")
	
	// Если ключ не установлен, генерируем новый ключ (только для разработки)
	if keyStr == "" {
		key := make([]byte, 32) // 256-битный ключ
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate JWT key: %v", err)
		}
		keyStr = base64.URLEncoding.EncodeToString(key)
		// Вывод предупреждения о использовании сгенерированного ключа (не должно произойти в производстве)
		fmt.Println("WARNING: JWT_SECRET_KEY not set, using a generated key. " +
			"Set JWT_SECRET_KEY environment variable in production.")
	}

	// Удаление любых пробелов и новой строки из ключа
	keyStr = strings.TrimSpace(keyStr)
	
	// Проверка длины ключа (рекомендуется минимум 32 байта для HS256)
	if len(keyStr) < 32 {
		return fmt.Errorf("JWT key is too short, must be at least 32 bytes")
	}

	JWTKey = []byte(keyStr)
	return nil
}

// Claims содержит пользовательские поля JWT-токена
type Claims struct {
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	IsRefresh  bool   `json:"is_refresh,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair представляет пару токенов доступа и обновления
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"` // в секундах
}

// GenerateJWT создает новый JWT-токен с указанными данными пользователя
func GenerateJWT(userID int, username, role string) (*TokenPair, error) {
	// Генерация токена доступа
	accessToken, _, err := generateToken(userID, username, role, AccessTokenExpiry, false)
	if err != nil {
		return nil, err
	}

	// Генерация токена обновления
	refreshToken, _, err := generateToken(userID, username, role, RefreshTokenExpiry, true)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenExpiry.Seconds()),
	}, nil
}

// generateToken генерирует JWT-токен с указанными параметрами
func generateToken(userID int, username, role string, expiry time.Duration, isRefresh bool) (string, time.Time, error) {
	// Установка времени истечения токена
	expirationTime := time.Now().Add(expiry)

	// Создание претензий с данными пользователя
	claims := &Claims{
		UserID:     userID,
		Username:   username,
		Role:       role,
		IsRefresh:  isRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "cinematique",
			Subject:   "user_auth",
		},
	}

	// Если это токен обновления, добавляем пользовательскую претензию
	if isRefresh {
		claims.RegisteredClaims.ID = fmt.Sprintf("%d_%d", userID, time.Now().UnixNano())
	}

	// Создание токена с методом подписи HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подпись токена
	tokenString, err := token.SignedString(JWTKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ParseJWT парсит и проверяет JWT-токен и возвращает претензии
func ParseJWT(tokenString string) (*Claims, error) {
	// Парсинг токена
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Проверка токена и возврат претензий
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateToken является псевдонимом для ParseJWT для обратной совместимости
func ValidateToken(tokenString string) (*Claims, error) {
	return ParseJWT(tokenString)
}
