// Package main demonstrates how to use the Keycloak integration package.
// This example shows various ways to work with Keycloak clients, managers, and factories.
//
// Usage:
//
//	go run examples/keycloak_example.go
//
// Note: This example requires a running Keycloak server for full functionality.
// Without a server, it will demonstrate the API usage but fail on actual token validation.
package main

func main() {
	keycloakMain()
}

import (
	"fmt"
	"log"

	"cinematique/internal/keycloak"
)

func keycloakMain() {
	// Пример использования Keycloak клиента
	config := keycloak.Config{
		ServerURL: "http://localhost:8080",
		Realm:     "cinematique",
		ClientID:  "cinematique-api",
	}

	// Создание клиента
	client := keycloak.NewClient(config)

	// Инициализация (загрузка JWKS)
	if err := client.Initialize(); err != nil {
		log.Fatalf("Failed to initialize Keycloak client: %v", err)
	}

	// Пример токена (в реальности получается от клиента)
	exampleToken := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJyc2EtZ2VuZXJhdGVkIn0..."

	// Проверка, является ли токен Keycloak токеном
	if client.IsKeycloakToken(exampleToken) {
		fmt.Println("This is a Keycloak token")

		// Валидация токена
		userInfo, err := client.ValidateTokenWithOptions(exampleToken, keycloak.DefaultValidationOptions())
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			return
		}

		// Вывод информации о пользователе
		fmt.Printf("User ID: %s\n", userInfo.ID)
		fmt.Printf("Username: %s\n", userInfo.Username)
		fmt.Printf("Email: %s\n", userInfo.Email)
		fmt.Printf("Roles: %v\n", userInfo.Roles)
		fmt.Printf("Local Role: %s\n", userInfo.LocalRole)
	} else {
		fmt.Println("This is not a Keycloak token")
	}

	// Пример использования менеджера
	manager := keycloak.GetGlobalManager()
	if err := manager.Initialize(config, true); err != nil {
		log.Printf("Failed to initialize manager: %v", err)
		return
	}

	if manager.IsEnabled() {
		fmt.Println("Keycloak is enabled")
		defaultClient := manager.GetDefaultClient()
		if defaultClient != nil {
			fmt.Println("Default client is available")
		}
	}

	// Пример использования фабрики
	factory := keycloak.GetGlobalFactory()
	client2, err := factory.GetOrCreateClient(config)
	if err != nil {
		log.Printf("Failed to get client from factory: %v", err)
		return
	}

	fmt.Printf("Client from factory: %v\n", client2 != nil)

	// Получение всех клиентов
	allClients := factory.GetAllClients()
	fmt.Printf("Total clients in factory: %d\n", len(allClients))
}
