package keycloak

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config.ServerURL, client.config.ServerURL)
	assert.Equal(t, config.Realm, client.config.Realm)
	assert.Equal(t, config.ClientID, client.config.ClientID)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestIsKeycloakToken(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

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
			result := client.IsKeycloakToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapKeycloakRoleToLocal(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	tests := []struct {
		name     string
		claims   *KeycloakClaims
		expected string
	}{
		{
			name: "admin role",
			claims: &KeycloakClaims{
				RealmAccess: struct {
					Roles []string `json:"roles"`
				}{
					Roles: []string{"admin"},
				},
			},
			expected: "admin",
		},
		{
			name: "user role",
			claims: &KeycloakClaims{
				RealmAccess: struct {
					Roles []string `json:"roles"`
				}{
					Roles: []string{"user"},
				},
			},
			expected: "user",
		},
		{
			name: "no roles - default to user",
			claims: &KeycloakClaims{
				RealmAccess: struct {
					Roles []string `json:"roles"`
				}{
					Roles: []string{},
				},
			},
			expected: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.MapKeycloakRoleToLocal(tt.claims)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserRoles(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	claims := &KeycloakClaims{
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{
			Roles: []string{"realm-role1", "realm-role2"},
		},
		ResourceAccess: map[string]struct {
			Roles []string `json:"roles"`
		}{
			"test-client": {
				Roles: []string{"client-role1", "client-role2"},
			},
		},
	}

	roles := client.GetUserRoles(claims)

	expected := []string{"realm-role1", "realm-role2", "client-role1", "client-role2"}
	assert.ElementsMatch(t, expected, roles)
}

func TestHasRole(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	claims := &KeycloakClaims{
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{
			Roles: []string{"admin", "user"},
		},
	}

	assert.True(t, client.HasRole(claims, "admin"))
	assert.True(t, client.HasRole(claims, "user"))
	assert.False(t, client.HasRole(claims, "nonexistent"))
}

func TestGetUserID(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	claims := &KeycloakClaims{
		Sub: "user-123",
	}

	userID := client.GetUserID(claims)
	assert.Equal(t, "user-123", userID)
}

func TestGetUsername(t *testing.T) {
	config := Config{
		ServerURL: "https://keycloak.example.com",
		Realm:     "test-realm",
		ClientID:  "test-client",
	}

	client := NewClient(config)

	tests := []struct {
		name     string
		claims   *KeycloakClaims
		expected string
	}{
		{
			name: "preferred username available",
			claims: &KeycloakClaims{
				PreferredUsername: "testuser",
				Email:             "test@example.com",
			},
			expected: "testuser",
		},
		{
			name: "no preferred username, use email",
			claims: &KeycloakClaims{
				PreferredUsername: "",
				Email:             "test@example.com",
			},
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.GetUsername(tt.claims)
			assert.Equal(t, tt.expected, result)
		})
	}
}
