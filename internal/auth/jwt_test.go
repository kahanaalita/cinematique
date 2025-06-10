package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseJWT(t *testing.T) {
	// Save original JWT key and restore it after the test
	originalKey := make([]byte, len(JWTKey))
	copy(originalKey, JWTKey)
	defer func() { JWTKey = originalKey }()

	// Set a test key
	JWTKey = []byte("test_secret_key")

	// Test data
	userID := 123
	username := "testuser"
	role := "user"

	// Test GenerateJWT
	tokenStr, err := GenerateJWT(userID, username, role)
	assert.NoError(t, err, "GenerateJWT should not return an error")
	assert.NotEmpty(t, tokenStr, "Generated token should not be empty")

	// Test ParseJWT
	claims, err := ParseJWT(tokenStr)
	assert.NoError(t, err, "ParseJWT should not return an error for a valid token")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, username, claims.Username, "Username should match")
	assert.Equal(t, role, claims.Role, "Role should match")
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 5*time.Second, "Expiration time should be ~24h from now")
}

func TestParseJWT_InvalidToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		expectErr error
	}{
		{
			name:      "empty token",
			token:     "",
			expectErr: jwt.ErrTokenMalformed,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			expectErr: jwt.ErrTokenMalformed,
		},
		{
			name:      "invalid signature",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInVzZXJuYW1lIjoidGVzdHVzZXIiLCJyb2xlIjoidXNlciIsImV4cCI6MTYxNzI5Mzk5OSwiaXNzIjoiY2luZW1hdGlxdWUiLCJzdWIiOiJ1c2VyX2F1dGgifQ.invalid_signature",
			expectErr: jwt.ErrTokenMalformed, // The token is malformed due to invalid signature format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseJWT(tt.token)
			assert.ErrorIs(t, err, tt.expectErr, "Expected error %v, got %v", tt.expectErr, err)
		})
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	// Create an expired token
	claims := &Claims{
		UserID:   123,
		Username: "testuser",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "cinematique",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWTKey)
	assert.NoError(t, err, "Failed to sign token")

	_, err = ParseJWT(tokenStr)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired, "Should return token expired error")
}

func TestJWT_InvalidSigningMethod(t *testing.T) {
	// Create a token with invalid signing method
	claims := &Claims{
		UserID:   123,
		Username: "testuser",
		Role:     "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err, "Failed to sign token with none method")

	_, err = ParseJWT(tokenStr)
	assert.ErrorIs(t, err, jwt.ErrSignatureInvalid, "Should return invalid signature error for none algorithm")
}
