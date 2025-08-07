package auth

import (
	"cinematique/internal/keycloak"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HybridAuthMiddleware поддерживает как JWT, так и Keycloak токены
func HybridAuthMiddleware(keycloakClient keycloak.KeycloakClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "отсутствует или неверный токен"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		// Сначала проверяем, является ли токен Keycloak токеном
		if keycloakClient != nil && keycloakClient.IsKeycloakToken(tokenStr) {
			userInfo, err := keycloakClient.ValidateTokenWithOptions(tokenStr, keycloak.DefaultValidationOptions())
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "неверный Keycloak токен"})
				return
			}

			// Устанавливаем контекст для Keycloak токена
			c.Set("auth_type", "keycloak")
			c.Set("user_id", userInfo.ID)
			c.Set("username", userInfo.Username)
			c.Set("role", userInfo.LocalRole)
			c.Set("user_info", userInfo)
			c.Next()
			return
		}

		// Если не Keycloak токен, пробуем как обычный JWT
		claims, err := ParseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
			return
		}

		// Проверяем, что токен не является refresh-токеном
		if claims.IsRefresh {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "refresh-токен не может быть использован для аутентификации"})
			return
		}

		// Устанавливаем контекст для обычного JWT токена
		c.Set("auth_type", "jwt")
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("jwt_claims", claims)
		c.Next()
	}
}

// JWTAuthMiddleware оставляем для обратной совместимости
func JWTAuthMiddleware() gin.HandlerFunc {
	return HybridAuthMiddleware(nil)
}

// OnlyAdminOrReadOnly разрешает только GET/OPTIONS/HEAD для обычных пользователей, а все методы — для админов
func OnlyAdminOrReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "нет роли в токене"})
			return
		}
		if role == "admin" {
			c.Next()
			return
		}
		// user: разрешаем только GET/OPTIONS/HEAD
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "только администратор может изменять данные"})
	}
}

// RequireRole требует определенную роль для доступа
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "нет роли в токене"})
			return
		}

		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("требуется роль %s", requiredRole)})
			return
		}

		c.Next()
	}
}

// RequireAnyRole требует любую из указанных ролей
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "нет роли в токене"})
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "неверный формат роли"})
			return
		}

		for _, role := range roles {
			if userRoleStr == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("требуется одна из ролей: %v", roles)})
	}
}

// GetUserFromContext извлекает информацию о пользователе из контекста
func GetUserFromContext(c *gin.Context) (*UserContext, error) {
	authType, exists := c.Get("auth_type")
	if !exists {
		return nil, fmt.Errorf("auth_type not found in context")
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	userContext := &UserContext{
		AuthType: authType.(string),
		UserID:   userID.(string),
		Username: username.(string),
		Role:     role.(string),
	}

	// Добавляем дополнительную информацию в зависимости от типа аутентификации
	switch authType {
	case "keycloak":
		if userInfo, exists := c.Get("user_info"); exists {
			userContext.KeycloakUserInfo = userInfo.(*keycloak.UserInfo)
		}
	case "jwt":
		if claims, exists := c.Get("jwt_claims"); exists {
			userContext.JWTClaims = claims.(*Claims)
		}
	}

	return userContext, nil
}

// UserContext содержит информацию о пользователе из контекста
type UserContext struct {
	AuthType         string             `json:"auth_type"`
	UserID           string             `json:"user_id"`
	Username         string             `json:"username"`
	Role             string             `json:"role"`
	KeycloakUserInfo *keycloak.UserInfo `json:"keycloak_user_info,omitempty"`
	JWTClaims        *Claims            `json:"jwt_claims,omitempty"`
}
