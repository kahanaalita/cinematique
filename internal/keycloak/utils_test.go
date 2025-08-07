package keycloak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTokenFormat(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{
			name:      "valid format",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectErr: false,
		},
		{
			name:      "invalid format - only 2 parts",
			token:     "header.payload",
			expectErr: true,
		},
		{
			name:      "invalid format - 4 parts",
			token:     "header.payload.signature.extra",
			expectErr: true,
		},
		{
			name:      "empty token",
			token:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenFormat(tt.token)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsKeycloakTokenByIssuer(t *testing.T) {
	// Создаем тестовый токен с известным issuer
	// Это упрощенный тест, в реальности нужен валидный JWT
	serverURL := "https://keycloak.example.com"
	realm := "test-realm"

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "invalid token format",
			token:    "invalid.token",
			expected: false,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeycloakTokenByIssuer(tt.token, serverURL, realm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRolesFromClaims(t *testing.T) {
	clientID := "test-client"

	claims := map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"realm-admin", "realm-user"},
		},
		"resource_access": map[string]interface{}{
			"test-client": map[string]interface{}{
				"roles": []interface{}{"client-admin", "client-user"},
			},
		},
	}

	roles := ExtractRolesFromClaims(claims, clientID)

	expected := []string{"realm-admin", "realm-user", "client-admin", "client-user"}
	assert.ElementsMatch(t, expected, roles)
}

func TestExtractRolesFromClaims_EmptyClaims(t *testing.T) {
	clientID := "test-client"
	claims := map[string]interface{}{}

	roles := ExtractRolesFromClaims(claims, clientID)

	assert.Empty(t, roles)
}

func TestSanitizeUserInfo(t *testing.T) {
	userInfo := &UserInfo{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Name:     "Test User",
		Roles:    []string{"user"},
	}

	sanitized := SanitizeUserInfo(userInfo)

	assert.NotNil(t, sanitized)
	assert.Equal(t, userInfo.ID, sanitized.ID)
	assert.Equal(t, userInfo.Username, sanitized.Username)
	assert.Equal(t, userInfo.Email, sanitized.Email)
}

func TestSanitizeUserInfo_Nil(t *testing.T) {
	result := SanitizeUserInfo(nil)
	assert.Nil(t, result)
}
