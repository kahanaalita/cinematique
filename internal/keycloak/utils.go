package keycloak

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DecodeTokenPayload декодирует payload JWT токена без проверки подписи
func DecodeTokenPayload(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Декодируем payload (вторая часть)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token payload: %w", err)
	}

	return claims, nil
}

// IsTokenExpired проверяет, истек ли токен
func IsTokenExpired(tokenString string) (bool, error) {
	claims, err := DecodeTokenPayload(tokenString)
	if err != nil {
		return true, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return true, fmt.Errorf("missing or invalid exp claim")
	}

	return int64(exp) < getCurrentTimestamp(), nil
}

// GetTokenIssuer извлекает issuer из токена
func GetTokenIssuer(tokenString string) (string, error) {
	claims, err := DecodeTokenPayload(tokenString)
	if err != nil {
		return "", err
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid iss claim")
	}

	return iss, nil
}

// GetTokenSubject извлекает subject из токена
func GetTokenSubject(tokenString string) (string, error) {
	claims, err := DecodeTokenPayload(tokenString)
	if err != nil {
		return "", err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid sub claim")
	}

	return sub, nil
}

// IsKeycloakTokenByIssuer проверяет, является ли токен токеном Keycloak по issuer
func IsKeycloakTokenByIssuer(tokenString, expectedServerURL, expectedRealm string) bool {
	issuer, err := GetTokenIssuer(tokenString)
	if err != nil {
		return false
	}

	expectedIssuer := fmt.Sprintf("%s/realms/%s", expectedServerURL, expectedRealm)
	return issuer == expectedIssuer
}

// ExtractRolesFromClaims извлекает роли из claims
func ExtractRolesFromClaims(claims map[string]interface{}, clientID string) []string {
	var roles []string

	// Извлекаем роли из realm_access
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if realmRoles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range realmRoles {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	// Извлекаем роли из resource_access для конкретного клиента
	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		if clientAccess, ok := resourceAccess[clientID].(map[string]interface{}); ok {
			if clientRoles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, role := range clientRoles {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	return roles
}

// getCurrentTimestamp возвращает текущий timestamp в секундах
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// ValidateTokenFormat проверяет базовый формат JWT токена
func ValidateTokenFormat(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Проверяем, что каждая часть может быть декодирована
	for i, part := range parts {
		if _, err := base64.RawURLEncoding.DecodeString(part); err != nil {
			return fmt.Errorf("invalid JWT part %d: %w", i, err)
		}
	}

	return nil
}

// SanitizeUserInfo очищает пользовательскую информацию от чувствительных данных
func SanitizeUserInfo(userInfo *UserInfo) *UserInfo {
	if userInfo == nil {
		return nil
	}

	sanitized := *userInfo
	// Можно добавить логику очистки чувствительных данных
	// Например, маскировать email или другие PII данные

	return &sanitized
}
