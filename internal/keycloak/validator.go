package keycloak

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

// TokenValidator интерфейс для валидации токенов
type TokenValidator interface {
	ValidateToken(tokenString string, options ValidationOptions) (*UserInfo, error)
	ValidateTokenWithClaims(tokenString string, options ValidationOptions) (*KeycloakClaims, *UserInfo, error)
}

// ValidateTokenWithOptions валидирует токен с дополнительными опциями
func (c *Client) ValidateTokenWithOptions(tokenString string, options ValidationOptions) (*UserInfo, error) {
	claims, userInfo, err := c.ValidateTokenWithClaims(tokenString, options)
	if err != nil {
		return nil, err
	}

	// Проверяем требуемые роли, если они указаны
	if len(options.RequiredRoles) > 0 {
		hasRequiredRole := false
		userRoles := c.GetUserRoles(claims)

		for _, requiredRole := range options.RequiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			return nil, fmt.Errorf("user does not have required roles: %v", options.RequiredRoles)
		}
	}

	return userInfo, nil
}

// ValidateTokenWithClaims валидирует токен и возвращает как claims, так и UserInfo
func (c *Client) ValidateTokenWithClaims(tokenString string, options ValidationOptions) (*KeycloakClaims, *UserInfo, error) {
	if c.keySet == nil {
		return nil, nil, ErrJWKSNotFetched
	}

	// Парсим токен с использованием JWKS
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(c.keySet))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	// Проверяем issuer, если требуется
	if options.ValidateIssuer {
		expectedIssuer := fmt.Sprintf("%s/realms/%s", c.config.ServerURL, c.config.Realm)
		if token.Issuer() != expectedIssuer {
			return nil, nil, fmt.Errorf("%w: expected %s, got %s", ErrInvalidIssuer, expectedIssuer, token.Issuer())
		}
	}

	// Проверяем expiration, если не разрешены истекшие токены
	if !options.AllowExpired && time.Now().After(token.Expiration()) {
		return nil, nil, ErrTokenExpired
	}

	// Проверяем audience, если требуется
	if options.ValidateAudience {
		audiences := token.Audience()
		validAudience := false
		for _, aud := range audiences {
			if aud == c.config.ClientID {
				validAudience = true
				break
			}
		}
		if !validAudience && len(audiences) > 0 {
			return nil, nil, fmt.Errorf("%w: expected %s in %v", ErrInvalidAudience, c.config.ClientID, audiences)
		}
	}

	// Извлекаем claims
	claims, err := c.extractClaims(token)
	if err != nil {
		return nil, nil, err
	}

	// Создаем UserInfo
	userInfo := &UserInfo{
		ID:                c.GetUserID(claims),
		Username:          c.GetUsername(claims),
		Email:             claims.Email,
		Name:              claims.Name,
		Roles:             c.GetUserRoles(claims),
		LocalRole:         c.MapKeycloakRoleToLocal(claims),
		PreferredUsername: claims.PreferredUsername,
		IsEmailVerified:   c.isEmailVerified(claims),
	}

	return claims, userInfo, nil
}

// extractClaims извлекает claims из JWT токена
func (c *Client) extractClaims(token jwt.Token) (*KeycloakClaims, error) {
	claims := &KeycloakClaims{}
	tokenMap, err := token.AsMap(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to convert token to map: %w", err)
	}

	// Маппим основные поля
	if sub, ok := tokenMap["sub"].(string); ok {
		claims.Sub = sub
	}
	if iss, ok := tokenMap["iss"].(string); ok {
		claims.Iss = iss
	}
	if exp, ok := tokenMap["exp"].(float64); ok {
		claims.Exp = int64(exp)
	}
	if iat, ok := tokenMap["iat"].(float64); ok {
		claims.Iat = int64(iat)
	}
	if username, ok := tokenMap["preferred_username"].(string); ok {
		claims.PreferredUsername = username
	}
	if email, ok := tokenMap["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := tokenMap["name"].(string); ok {
		claims.Name = name
	}
	claims.Aud = tokenMap["aud"]

	// Маппим realm_access
	if realmAccess, ok := tokenMap["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					claims.RealmAccess.Roles = append(claims.RealmAccess.Roles, roleStr)
				}
			}
		}
	}

	// Маппим resource_access
	if resourceAccess, ok := tokenMap["resource_access"].(map[string]interface{}); ok {
		claims.ResourceAccess = make(map[string]struct {
			Roles []string `json:"roles"`
		})

		for clientID, clientData := range resourceAccess {
			if clientMap, ok := clientData.(map[string]interface{}); ok {
				if roles, ok := clientMap["roles"].([]interface{}); ok {
					var roleStrings []string
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							roleStrings = append(roleStrings, roleStr)
						}
					}
					claims.ResourceAccess[clientID] = struct {
						Roles []string `json:"roles"`
					}{Roles: roleStrings}
				}
			}
		}
	}

	return claims, nil
}

// isEmailVerified проверяет, подтвержден ли email
func (c *Client) isEmailVerified(claims *KeycloakClaims) bool {
	// Эта информация может быть в дополнительных claims
	// Пока возвращаем true по умолчанию
	return true
}

// GetTokenInfo возвращает информацию о токене
func (c *Client) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	token, err := jwt.Parse([]byte(tokenString), jwt.WithVerify(false))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	info := &TokenInfo{
		Type:      TokenTypeAccess, // По умолчанию считаем access токеном
		Subject:   token.Subject(),
		Issuer:    token.Issuer(),
		ExpiresAt: token.Expiration().Unix(),
		IssuedAt:  token.IssuedAt().Unix(),
		Valid:     time.Now().Before(token.Expiration()),
	}

	// Определяем audience
	audiences := token.Audience()
	if len(audiences) > 0 {
		info.Audience = audiences[0]
	}

	return info, nil
}
