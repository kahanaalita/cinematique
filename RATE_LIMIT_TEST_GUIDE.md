# Полное руководство по тестированию Rate Limiting в Cinematique API

## Обзор
Это руководство предоставляет полный набор команд и скриптов для тестирования rate limiting в приложении Cinematique через curl запросы.

## Предварительные требования
1. Запущенное приложение (main.go)
2. Запущенные сервисы (docker-compose)
3. Установленные утилиты: curl, jq

## Быстрый старт

### 1. Регистрация и авторизация
```bash
# Регистрация нового пользователя
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "testpass123", "role": "user"}'

# Авторизация и получение токена
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}' | jq -r '.access_token')
```

### 2. Базовые тесты rate limiting

#### Проверка статуса rate limit
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/rate-limit/status?endpoint=/api/movies"
```

#### Проверка заголовков rate limit
```bash
curl -s -i -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/movies | head -10
```

### 3. Нагрузочное тестирование

#### Медленное нагрузочное тестирование (15 запросов)
```bash
for i in {1..15}; do
  echo "Запрос $i"
  curl -s -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/movies | jq -r '.movies | length'
  sleep 0.5
done
```

#### Быстрое нагрузочное тестирование (для проверки 429)
```bash
# Создание 50 быстрых запросов
for i in {1..50}; do
  curl -s -w "%{http_code} " -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/movies > /dev/null &
done
wait
echo "Запросы отправлены"
```

### 4. Проверка различных endpoint

#### Movies endpoint
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/movies
```

#### Actors endpoint
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/actors
```

#### С параметрами
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/movies?page=1&limit=5"
```

### 5. Проверка 429 ошибки

#### Массовые запросы для получения 429
```bash
# Создание 1005 запросов для превышения лимита
for i in {1..1005}; do
  response=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/movies)
  
  http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d':' -f2)
  echo "Запрос $i: HTTP $http_code"
  
  if [ "$http_code" = "429" ]; then
    echo "Rate limit достигнут на запросе $i!"
    echo "Ответ: $(echo "$response" | head -5)"
    break
  fi
  
  # Небольшая задержка для предотвращения блокировки
  sleep 0.1
done
```

### 6. Мониторинг через Redis

#### Проверка Redis ключей
```bash
# Просмотр текущих rate limit ключей
redis-cli -p 6379 keys "*ratelimit*"

# Просмотр значений ключей
redis-cli -p 6379 get "ratelimit:4:::1:202507241658"
```

## Полезные скрипты

### Автоматический тест
```bash
#!/bin/bash
# Сохранить как test_rate_limit_auto.sh

TOKEN="your_token_here"
BASE_URL="http://localhost:8080"

echo "=== Автоматическое тестирование Rate Limiting ==="

# 1. Проверка начального состояния
status=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
echo "Initial: $(echo $status | jq '{current_count, remaining}')"

# 2. 10 последовательных запросов
for i in {1..10}; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/movies" > /dev/null
  
  status=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
  count=$(echo $status | jq -r '.current_count')
  echo "Request $i: count=$count"
done

echo "Тест завершен!"
```

### Проверка заголовков
```bash
#!/bin/bash
# Сохранить как check_headers.sh

TOKEN="your_token_here"

curl -s -i -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/movies | \
  grep -E "(X-RateLimit|HTTP)" | \
  sed 's/^/  /'
```

## Ожидаемые результаты

### Успешные запросы (200 OK)
```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1706097600
```

### Превышение лимита (429 Too Many Requests)
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1706097600

{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Maximum 1000 requests per 1m0s",
  "current_count": 1001,
  "limit": 1000,
  "window": "1m0s",
  "retry_after": 60
}
```

## Примеры команд для разных сценариев

### 1. Тестирование с разными пользователями
```bash
# Создание нескольких пользователей
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "email": "user1@example.com", "password": "pass123", "role": "user"}'

curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user2", "email": "user2@example.com", "password": "pass123", "role": "user"}'
```

### 2. Тестирование с разными IP
```bash
# Использование разных IP через прокси или VPN
# Rate limit работает по комбинации user_id + IP + endpoint
```

### 3. Проверка разных endpoint
```bash
# Проверка всех защищенных endpoint
endpoints=(
  "/api/movies"
  "/api/movies/1"
  "/api/actors"
  "/api/actors/1"
  "/api/movies/1/actors"
)

for endpoint in "${endpoints[@]}"; do
  echo "Testing $endpoint"
  curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8080$endpoint" | jq '.error // .'
done
```

## Диагностика проблем

### Проверка логов
```bash
# Просмотр логов приложения
tail -f app.log | grep -i "rate\|limit"

# Проверка Redis
redis-cli -p 6379 monitor
```

### Проверка конфигурации
```bash
# Проверка переменных окружения
env | grep -i "rate\|redis"

# Проверка доступности Redis
redis-cli -p 6379 ping
```

## Заключение

Rate limiting в Cinematique API успешно протестирован и работает корректно:
- ✅ Счетчик запросов увеличивается
- ✅ Заголовки X-RateLimit присутствуют
- ✅ 429 ошибка возвращается при превышении лимита
- ✅ Различные endpoint имеют rate limiting
- ✅ Статус endpoint работает для мониторинга
