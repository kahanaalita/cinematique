package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Config содержит конфигурацию для Keycloak
type Config struct {
	ServerURL string `json:"server_url"`
	Realm     string `json:"realm"`
	ClientID  string `json:"client_id"`
}

// Client представляет клиент для работы с Keycloak
type Client struct {
	config     Config
	keySet     jwk.Set
	httpClient *http.Client
}

// KeycloakClaims представляет claims из Keycloak JWT токена
type KeycloakClaims struct {
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	PreferredUsername string      `json:"preferred_username"`
	Email             string      `json:"email"`
	Name              string      `json:"name"`
	Sub               string      `json:"sub"` // Subject (user ID)
	Iss               string      `json:"iss"` // Issuer
	Aud               interface{} `json:"aud"` // Audience
	Exp               int64       `json:"exp"` // Expiration
	Iat               int64       `json:"iat"` // Issued at
}

// NewClient создает новый клиент Keycloak
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Initialize инициализирует клиент, получая публичные ключи от Keycloak
func (c *Client) Initialize() error {
	return c.fetchJWKS()
}

// fetchJWKS получает JWKS от Keycloak
func (c *Client) fetchJWKS() error {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid_connect/certs",
		c.config.ServerURL, c.config.Realm)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	keySet, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	c.keySet = keySet
	return nil
}

// ValidateToken проверяет Keycloak JWT токен с стандартными опциями
func (c *Client) ValidateToken(tokenString string) (*KeycloakClaims, error) {
	claims, _, err := c.ValidateTokenWithClaims(tokenString, DefaultValidationOptions())
	return claims, err
}

// IsKeycloakToken проверяет, является ли токен токеном Keycloak
func (c *Client) IsKeycloakToken(tokenString string) bool {
	// Простая проверка - пытаемся декодировать header
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return false
	}

	// Парсим токен без проверки подписи для получения issuer
	token, err := jwt.Parse([]byte(tokenString), jwt.WithVerify(false))
	if err != nil {
		return false
	}

	// Проверяем issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", c.config.ServerURL, c.config.Realm)
	return token.Issuer() == expectedIssuer
}

// GetUserRoles извлекает роли пользователя из claims
func (c *Client) GetUserRoles(claims *KeycloakClaims) []string {
	var roles []string

	// Добавляем роли из realm_access
	roles = append(roles, claims.RealmAccess.Roles...)

	// Добавляем роли из resource_access для нашего клиента
	if clientRoles, exists := claims.ResourceAccess[c.config.ClientID]; exists {
		roles = append(roles, clientRoles.Roles...)
	}

	return roles
}

// HasRole проверяет, есть ли у пользователя определенная роль
func (c *Client) HasRole(claims *KeycloakClaims, role string) bool {
	roles := c.GetUserRoles(claims)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// MapKeycloakRoleToLocal маппит роли Keycloak на локальные роли
func (c *Client) MapKeycloakRoleToLocal(claims *KeycloakClaims) string {
	roles := c.GetUserRoles(claims)

	// Проверяем роли в порядке приоритета
	for _, role := range roles {
		switch role {
		case "admin", "administrator", "cinematique-admin":
			return "admin"
		case "user", "cinematique-user":
			return "user"
		}
	}

	// По умолчанию возвращаем роль пользователя
	return "user"
}

// GetUserID возвращает ID пользователя из claims
func (c *Client) GetUserID(claims *KeycloakClaims) string {
	return claims.Sub
}

// GetUsername возвращает имя пользователя из claims
func (c *Client) GetUsername(claims *KeycloakClaims) string {
	if claims.PreferredUsername != "" {
		return claims.PreferredUsername
	}
	return claims.Email
}
