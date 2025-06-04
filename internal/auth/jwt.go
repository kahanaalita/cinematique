package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// JWTKey содержит секретный ключ для подписи JWT токенов
var JWTKey = []byte("В продакшене замени на получение из переменных окружения")

// Claims содержит кастомные поля JWT токена
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT создает новый JWT токен с указанными данными пользователя
// и временем жизни 24 часа
func GenerateJWT(userID int, username, role string) (string, error) {
	// Устанавливаем время истечения токена
	expirationTime := time.Now().Add(24 * time.Hour)

	// Создаем claims с пользовательскими данными
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "cinematique",
			Subject:   "user_auth",
		},
	}

	// Создаем токен с алгоритмом подписи HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	return token.SignedString(JWTKey)
}

// ParseJWT проверяет и парсит JWT токен, возвращая claims
func ParseJWT(tokenStr string) (*Claims, error) {
	// Парсим токен с claims
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return JWTKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена и приводим claims к нашему типу
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
