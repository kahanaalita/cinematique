package keycloak

// TokenType определяет тип токена
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeID      TokenType = "id"
)

// UserInfo содержит информацию о пользователе из Keycloak
type UserInfo struct {
	ID                string   `json:"id"`
	Username          string   `json:"username"`
	Email             string   `json:"email"`
	Name              string   `json:"name"`
	Roles             []string `json:"roles"`
	LocalRole         string   `json:"local_role"`
	IsEmailVerified   bool     `json:"email_verified"`
	PreferredUsername string   `json:"preferred_username"`
}

// TokenInfo содержит информацию о токене
type TokenInfo struct {
	Type      TokenType `json:"type"`
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	Audience  string    `json:"audience"`
	ExpiresAt int64     `json:"expires_at"`
	IssuedAt  int64     `json:"issued_at"`
	Valid     bool      `json:"valid"`
}

// ValidationOptions опции для валидации токена
type ValidationOptions struct {
	ValidateAudience bool
	ValidateIssuer   bool
	RequiredRoles    []string
	AllowExpired     bool
}

// DefaultValidationOptions возвращает стандартные опции валидации
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		ValidateAudience: true,
		ValidateIssuer:   true,
		RequiredRoles:    nil,
		AllowExpired:     false,
	}
}
