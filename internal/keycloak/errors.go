package keycloak

import "errors"

// Определяем специфичные ошибки для Keycloak
var (
	ErrTokenInvalid     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrJWKSNotFetched   = errors.New("JWKS not initialized")
	ErrInvalidAudience  = errors.New("invalid audience")
	ErrMissingClaims    = errors.New("missing required claims")
	ErrClientNotEnabled = errors.New("keycloak client not enabled")
)
