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
	tokenPair, err := GenerateJWT(userID, username, role)
	assert.NoError(t, err, "GenerateJWT should not return an error")
	assert.NotNil(t, tokenPair, "TokenPair should not be nil")
	assert.NotEmpty(t, tokenPair.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, tokenPair.RefreshToken, "Refresh token should not be empty")

	// Test ParseJWT with access token
	claims, err := ParseJWT(tokenPair.AccessToken)
	assert.NoError(t, err, "ParseJWT should not return an error for a valid token")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, username, claims.Username, "Username should match")
	assert.Equal(t, role, claims.Role, "Role should match")
	assert.WithinDuration(t, time.Now().Add(AccessTokenExpiry), claims.ExpiresAt.Time, 5*time.Second, "Expiration time should match AccessTokenExpiry")

	// Test ParseJWT with refresh token
	refreshClaims, err := ParseJWT(tokenPair.RefreshToken)
	assert.NoError(t, err, "ParseJWT should not return an error for a valid refresh token")
	assert.Equal(t, userID, refreshClaims.UserID, "UserID should match in refresh token")
	assert.Equal(t, username, refreshClaims.Username, "Username should match in refresh token")
	assert.Equal(t, role, refreshClaims.Role, "Role should match in refresh token")
	assert.True(t, refreshClaims.IsRefresh, "Refresh token should have IsRefresh=true")
	assert.WithinDuration(t, time.Now().Add(RefreshTokenExpiry), refreshClaims.ExpiresAt.Time, 5*time.Second, "Expiration time should match RefreshTokenExpiry")
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
	// Check if the error contains the expected message about the signing method
	assert.ErrorContains(t, err, "unexpected signing method: none", "Should return error about unexpected signing method")
}
