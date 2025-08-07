package keycloak

import (
	"fmt"
	"sync"
)

// Manager управляет несколькими Keycloak клиентами
type Manager struct {
	factory       *ClientFactory
	defaultClient KeycloakClient
	enabled       bool
	mutex         sync.RWMutex
}

// NewManager создает новый менеджер Keycloak
func NewManager() *Manager {
	return &Manager{
		factory: NewClientFactory(),
		enabled: false,
	}
}

// Initialize инициализирует менеджер с конфигурацией по умолчанию
func (m *Manager) Initialize(config Config, enabled bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.enabled = enabled

	if !enabled {
		return nil
	}

	client, err := m.factory.GetOrCreateClient(config)
	if err != nil {
		return fmt.Errorf("failed to initialize default keycloak client: %w", err)
	}

	m.defaultClient = client
	return nil
}

// IsEnabled возвращает true, если Keycloak включен
func (m *Manager) IsEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.enabled
}

// GetDefaultClient возвращает клиент по умолчанию
func (m *Manager) GetDefaultClient() KeycloakClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.defaultClient
}

// GetClient возвращает клиент для указанной конфигурации
func (m *Manager) GetClient(config Config) (KeycloakClient, error) {
	if !m.IsEnabled() {
		return nil, ErrClientNotEnabled
	}

	return m.factory.GetOrCreateClient(config)
}

// ValidateTokenWithAnyClient пытается валидировать токен с любым доступным клиентом
func (m *Manager) ValidateTokenWithAnyClient(tokenString string) (*UserInfo, error) {
	if !m.IsEnabled() {
		return nil, ErrClientNotEnabled
	}

	clients := m.factory.GetAllClients()
	if len(clients) == 0 && m.defaultClient != nil {
		// Используем клиент по умолчанию
		return m.defaultClient.ValidateTokenWithOptions(tokenString, DefaultValidationOptions())
	}

	var lastErr error
	for _, client := range clients {
		if client.IsKeycloakToken(tokenString) {
			userInfo, err := client.ValidateTokenWithOptions(tokenString, DefaultValidationOptions())
			if err == nil {
				return userInfo, nil
			}
			lastErr = err
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("token not recognized by any keycloak client")
}

// IsKeycloakTokenByAnyClient проверяет, является ли токен токеном Keycloak для любого клиента
func (m *Manager) IsKeycloakTokenByAnyClient(tokenString string) bool {
	if !m.IsEnabled() {
		return false
	}

	// Сначала проверяем клиент по умолчанию
	if m.defaultClient != nil && m.defaultClient.IsKeycloakToken(tokenString) {
		return true
	}

	// Затем проверяем все остальные клиенты
	clients := m.factory.GetAllClients()
	for _, client := range clients {
		if client.IsKeycloakToken(tokenString) {
			return true
		}
	}

	return false
}

// Shutdown корректно завершает работу менеджера
func (m *Manager) Shutdown() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.enabled = false
	m.defaultClient = nil
	// Фабрика будет очищена сборщиком мусора
}

// Глобальный менеджер
var globalManager = NewManager()

// GetGlobalManager возвращает глобальный менеджер
func GetGlobalManager() *Manager {
	return globalManager
}

// InitializeGlobal инициализирует глобальный менеджер
func InitializeGlobal(config Config, enabled bool) error {
	return globalManager.Initialize(config, enabled)
}
