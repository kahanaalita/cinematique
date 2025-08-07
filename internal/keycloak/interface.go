package keycloak

// KeycloakClient интерфейс для работы с Keycloak
type KeycloakClient interface {
	// Initialize инициализирует клиент
	Initialize() error

	// Валидация токенов
	ValidateToken(tokenString string) (*KeycloakClaims, error)
	ValidateTokenWithOptions(tokenString string, options ValidationOptions) (*UserInfo, error)
	ValidateTokenWithClaims(tokenString string, options ValidationOptions) (*KeycloakClaims, *UserInfo, error)

	// Проверка токенов
	IsKeycloakToken(tokenString string) bool
	GetTokenInfo(tokenString string) (*TokenInfo, error)

	// Работа с ролями
	GetUserRoles(claims *KeycloakClaims) []string
	HasRole(claims *KeycloakClaims, role string) bool
	MapKeycloakRoleToLocal(claims *KeycloakClaims) string

	// Информация о пользователе
	GetUserID(claims *KeycloakClaims) string
	GetUsername(claims *KeycloakClaims) string
}

// Убеждаемся, что Client реализует интерфейс
var _ KeycloakClient = (*Client)(nil)
