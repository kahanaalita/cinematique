package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "отсутствует или неверный токен"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
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
		
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
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
