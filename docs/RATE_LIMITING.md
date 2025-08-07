# Rate Limiting в Cinematique API

## Обзор

Система rate limiting защищает API от избыточных запросов, используя Redis для хранения счетчиков и алгоритм Fixed Window Counter.

## Конфигурация

### Переменные окружения

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiting Configuration
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
RATE_LIMIT_WINDOW_SECONDS=60
```

### Настройки по умолчанию

- **Лимит**: 1000 запросов в минуту
- **Окно**: 60 секунд (фиксированное)
- **Ограничиваемые endpoints**: `/api/movies`, `/api/actors`

## Алгоритм работы

### Fixed Window Counter

1. **Ключ в Redis**: `ratelimit:{user_id}:{ip}:{endpoint}:{timestamp_minute}`
2. **Пример ключа**: `ratelimit:12345:192.168.0.1:/api/movies:202507241004`
3. **Логика**:
   - При каждом запросе увеличиваем счетчик (`INCR`)
   - Если первый запрос в окне, устанавливаем TTL (60 сек)
   - Если счетчик превышает лимит → возвращаем 429 Too Many Requests

### Идентификация пользователя

Rate limiting работает по комбинации:
- **user_id** (из JWT токена или Keycloak)
- **IP адрес** (для защиты от DDoS)
- **endpoint** (разные лимиты для разных API)

## HTTP Headers

### Response Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1706097600
```

### При превышении лимита (429)

```json
{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Maximum 1000 requests per 1m0s",
  "current_count": 1001,
  "limit": 1000,
  "window": "1m0s",
  "retry_after": 60
}
```

## Мониторинг

### Endpoint для проверки статуса

```http
GET /api/rate-limit/status?endpoint=/api/movies
```

**Response:**
```json
{
  "enabled": true,
  "user_id": "12345",
  "ip": "192.168.0.1",
  "endpoint": "/api/movies",
  "current_count": 45,
  "limit": 1000,
  "remaining": 955,
  "window": "1m0s",
  "reset_time": 1706097600,
  "reset_time_human": "2025-01-24T10:05:00Z",
  "restricted_endpoints": ["/api/movies", "/api/actors"]
}
```

## Развертывание

### Docker Compose

Redis автоматически запускается в docker-compose:

```yaml
redis:
  image: redis:7-alpine
  container_name: redis
  ports:
    - "6379:6379"
  command: redis-server --appendonly yes
  volumes:
    - redis_data:/data
```

### Запуск

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка Redis
docker-compose logs redis

# Запуск приложения
go run main.go
```

## Тестирование

### Проверка лимитов

```bash
# Проверка статуса
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/rate-limit/status

# Массовые запросы для тестирования лимита
for i in {1..1005}; do
  curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    http://localhost:8080/api/movies
done
```

### Unit тесты

```bash
go test ./internal/ratelimit/...
```

## Производительность

### Характеристики Redis

- **Скорость**: ~100,000 операций INCR в секунду
- **Память**: ~50 байт на ключ
- **TTL**: Автоматическое удаление устаревших ключей

### Масштабирование

- **Redis Cluster**: Для высоких нагрузок
- **Sharding**: По user_id или IP
- **Репликация**: Master-Slave для отказоустойчивости

## Безопасность

### Защита от обхода

1. **IP Detection**: Учитывает X-Forwarded-For, X-Real-IP
2. **User Identification**: JWT + Keycloak
3. **Endpoint Isolation**: Разные лимиты для разных API

### Рекомендации

- Используйте HTTPS для защиты токенов
- Настройте правильные заголовки прокси
- Мониторьте подозрительную активность

## Troubleshooting

### Частые проблемы

1. **Redis недоступен**:
   ```
   Failed to connect to Redis: dial tcp 127.0.0.1:6379: connect: connection refused
   ```
   **Решение**: Проверьте, запущен ли Redis

2. **Rate limiting не работает**:
   - Проверьте `RATE_LIMIT_ENABLED=true`
   - Убедитесь, что endpoint в списке `RestrictedEndpoints`

3. **Неправильный user_id**:
   - Проверьте JWT токен
   - Убедитесь в правильности middleware порядка

### Логи

```bash
# Включить debug логи Redis
export REDIS_LOG_LEVEL=debug

# Проверить логи приложения
tail -f app.log | grep "rate"
```

## Метрики

Rate limiting интегрирован с Prometheus:

- `http_requests_total` - общее количество запросов
- `http_request_duration_seconds` - время обработки
- Заголовки `X-RateLimit-*` для мониторинга

Дашборд Grafana включает графики rate limiting.