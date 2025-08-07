package keycloak

import (
	"fmt"
	"sync"
)

// ClientFactory фабрика для создания Keycloak клиентов
type ClientFactory struct {
	clients map[string]KeycloakClient
	mutex   sync.RWMutex
}

// NewClientFactory создает новую фабрику клиентов
func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		clients: make(map[string]KeycloakClient),
	}
}

// GetOrCreateClient получает существующий или создает новый клиент
func (f *ClientFactory) GetOrCreateClient(config Config) (KeycloakClient, error) {
	key := f.getClientKey(config)

	f.mutex.RLock()
	if client, exists := f.clients[key]; exists {
		f.mutex.RUnlock()
		return client, nil
	}
	f.mutex.RUnlock()

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Двойная проверка после получения блокировки записи
	if client, exists := f.clients[key]; exists {
		return client, nil
	}

	// Создаем новый клиент
	client := NewClient(config)
	if err := client.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize keycloak client: %w", err)
	}

	f.clients[key] = client
	return client, nil
}

// getClientKey создает уникальный ключ для клиента
func (f *ClientFactory) getClientKey(config Config) string {
	return fmt.Sprintf("%s:%s:%s", config.ServerURL, config.Realm, config.ClientID)
}

// RemoveClient удаляет клиент из фабрики
func (f *ClientFactory) RemoveClient(config Config) {
	key := f.getClientKey(config)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.clients, key)
}

// GetAllClients возвращает все активные клиенты
func (f *ClientFactory) GetAllClients() map[string]KeycloakClient {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	result := make(map[string]KeycloakClient)
	for key, client := range f.clients {
		result[key] = client
	}
	return result
}

// Глобальная фабрика клиентов
var globalFactory = NewClientFactory()

// GetGlobalFactory возвращает глобальную фабрику клиентов
func GetGlobalFactory() *ClientFactory {
	return globalFactory
}
