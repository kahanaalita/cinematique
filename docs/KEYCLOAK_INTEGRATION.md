# Keycloak Integration

Этот проект поддерживает гибридный подход аутентификации, где Keycloak работает параллельно с существующей JWT-системой.

## Архитектура

### Гибридный подход
- **JWT токены** - для внутренних пользователей и существующих интеграций
- **Keycloak токены** - для внешних клиентов и новых интеграций
- **Общий middleware** - автоматически определяет тип токена и обрабатывает соответствующим образом

### Компоненты

#### 1. Keycloak Client (`internal/keycloak/client.go`)
- Основной клиент для работы с Keycloak
- Валидация JWT токенов с использованием JWKS
- Извлечение ролей и пользовательской информации

#### 2. Factory Pattern (`internal/keycloak/factory.go`)
- Создание и управление клиентами Keycloak
- Кэширование клиентов для повышения производительности
- Thread-safe операции

#### 3. Manager (`internal/keycloak/manager.go`)
- Управление несколькими Keycloak клиентами
- Глобальная инициализация и конфигурация
- Graceful shutdown

#### 4. Hybrid Middleware (`internal/auth/middleware.go`)
- Автоматическое определение типа токена (JWT vs Keycloak)
- Единый интерфейс для обоих типов аутентификации
- Обратная совместимость с существующими JWT токенами

## Конфигурация

### Переменные окружения

```bash
# Включение/отключение Keycloak
KEYCLOAK_ENABLED=true

# URL сервера Keycloak
KEYCLOAK_SERVER_URL=http://localhost:8080

# Realm в Keycloak
KEYCLOAK_REALM=cinematique

# Client ID в Keycloak
KEYCLOAK_CLIENT_ID=cinematique-api
```

### Настройка Keycloak

1. **Создание Realm**
   ```
   Realm Name: cinematique
   ```

2. **Создание Client**
   ```
   Client ID: cinematique-api
   Client Protocol: openid-connect
   Access Type: public (для SPA) или confidential (для backend)
   ```

3. **Настройка ролей**
   ```
   Realm Roles:
   - admin
   - user
   
   Client Roles (cinematique-api):
   - cinematique-admin
   - cinematique-user
   ```

## Использование

### Инициализация

```go
// В main.go или cmd/app.go
cfg := config.LoadConfig()

// Инициализация Keycloak
if err := keycloak.InitializeGlobal(cfg.Keycloak.ToKeycloakConfig(), cfg.Keycloak.Enabled); err != nil {
    log.Printf("Failed to initialize Keycloak: %v", err)
}
```

### Middleware

```go
// Гибридный middleware (поддерживает JWT и Keycloak)
keycloakManager := keycloak.GetGlobalManager()
var keycloakClient keycloak.KeycloakClient
if keycloakManager.IsEnabled() {
    keycloakClient = keycloakManager.GetDefaultClient()
}

router.Use(auth.HybridAuthMiddleware(keycloakClient))
```

### Получение информации о пользователе

```go
func MyHandler(c *gin.Context) {
    // Получение контекста пользователя
    userContext, err := auth.GetUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user context"})
        return
    }

    // Проверка типа аутентификации
    switch userContext.AuthType {
    case "keycloak":
        // Работа с Keycloak пользователем
        keycloakInfo := userContext.KeycloakUserInfo
        log.Printf("Keycloak user: %s, roles: %v", keycloakInfo.Username, keycloakInfo.Roles)
    case "jwt":
        // Работа с JWT пользователем
        jwtClaims := userContext.JWTClaims
        log.Printf("JWT user: %s, role: %s", jwtClaims.Username, jwtClaims.Role)
    }
}
```

## Маппинг ролей

Keycloak роли автоматически маппятся на локальные роли:

| Keycloak Role | Local Role |
|---------------|------------|
| admin, administrator, cinematique-admin | admin |
| user, cinematique-user | user |
| (любая другая роль) | user |

## Безопасность

### Валидация токенов
- **JWT токены**: Проверка подписи с использованием HMAC-SHA256
- **Keycloak токены**: Проверка подписи с использованием RSA и JWKS
- **Общие проверки**: Expiration, Issuer, Audience

### Best Practices
1. Используйте HTTPS в продакшене
2. Регулярно обновляйте JWKS ключи
3. Настройте правильные Audience для клиентов
4. Используйте короткое время жизни для access токенов
5. Логируйте все попытки аутентификации

## Тестирование

### Запуск тестов
```bash
# Тесты Keycloak пакета
go test ./internal/keycloak/... -v

# Тесты middleware
go test ./internal/auth/... -v

# Все тесты
go test ./... -v
```

### Примеры токенов

#### JWT токен (внутренний)
```bash
curl -H "Authorization: Bearer <jwt-token>" http://localhost:8080/api/movies
```

#### Keycloak токен (внешний)
```bash
curl -H "Authorization: Bearer <keycloak-token>" http://localhost:8080/api/movies
```

## Мониторинг

### Метрики
- Количество успешных аутентификаций по типам
- Количество неудачных попыток аутентификации
- Время отклика JWKS endpoint

### Логирование
- Инициализация Keycloak клиентов
- Ошибки валидации токенов
- Переключение между типами аутентификации

## Troubleshooting

### Частые проблемы

1. **JWKS не загружается**
   - Проверьте доступность Keycloak сервера
   - Убедитесь в правильности URL и realm

2. **Токен не валидируется**
   - Проверьте issuer в токене
   - Убедитесь в правильности audience
   - Проверьте время жизни токена

3. **Роли не маппятся**
   - Проверьте настройки ролей в Keycloak
   - Убедитесь в правильности client_id

### Отладка

Включите детальное логирование:
```go
log.SetLevel(log.DebugLevel)
```

Проверьте конфигурацию:
```bash
curl http://localhost:8080/realms/cinematique/.well-known/openid_connect/configuration
```