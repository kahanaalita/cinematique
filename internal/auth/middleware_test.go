package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestJWTAuthMiddleware(t *testing.T) {
	// Save original JWT key and restore it after the test
	originalKey := make([]byte, len(JWTKey))
	copy(originalKey, JWTKey)
	defer func() { JWTKey = originalKey }()

	// Set a test key
	JWTKey = []byte("test_secret_key")

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectedBody   string
		shouldSetUser bool
	}{
		{
			name: "valid token",
			setupRequest: func() *http.Request {
				tokenPair, _ := GenerateJWT(123, "testuser", "user")
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
				return req
			},
			expectedStatus: http.StatusOK,
			shouldSetUser: true,
		},
		{
			name: "missing authorization header",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/test", nil)
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"отсутствует или неверный токен"}`,
			shouldSetUser: false,
		},
		{
			name: "invalid token format",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "InvalidToken")
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"отсутствует или неверный токен"}`,
			shouldSetUser: false,
		},
		{
			name: "invalid token",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer invalid.token.here")
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"неверный токен"}`,
			shouldSetUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupRouter()
			var userID, username, role interface{}
			
			r.GET("/test", JWTAuthMiddleware(), func(c *gin.Context) {
				userID, _ = c.Get("user_id")
				username, _ = c.Get("username")
				role, _ = c.Get("role")
				c.Status(http.StatusOK)
			})

			// Execute
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.setupRequest())

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			if tt.shouldSetUser {
				assert.Equal(t, 123, userID)
				assert.Equal(t, "testuser", username)
				assert.Equal(t, "user", role)
			} else {
				nilValue := interface{}(nil)
				assert.Equal(t, nilValue, userID)
				assert.Equal(t, nilValue, username)
				assert.Equal(t, nilValue, role)
			}
		})
	}
}

func TestOnlyAdminOrReadOnly(t *testing.T) {
	tests := []struct {
		name           string
		method        string
		setupContext  func(c *gin.Context)
		expectedStatus int
		shouldAllow   bool
	}{
		{
			name:    "admin can POST",
			method:  "POST",
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
			},
			expectedStatus: http.StatusOK,
			shouldAllow:   true,
		},
		{
			name:    "user can GET",
			method:  "GET",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusOK,
			shouldAllow:   true,
		},
		{
			name:    "user cannot POST",
			method:  "POST",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
			shouldAllow:   false,
		},
		{
			name:    "missing role",
			method:  "GET",
			setupContext: func(c *gin.Context) {
				// Don't set any role
			},
			expectedStatus: http.StatusForbidden,
			shouldAllow:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			r := setupRouter()
			var handlerCalled bool
			
			r.Handle(tt.method, "/test", func(c *gin.Context) {
				tt.setupContext(c)
			}, OnlyAdminOrReadOnly(), func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})

			// Execute
			req, _ := http.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.shouldAllow, handlerCalled, "Handler called state mismatch")
		})
	}
}
