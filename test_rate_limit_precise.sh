#!/bin/bash

# Точное тестирование rate limiting с подсчетом запросов

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGUiOiJ1c2VyIiwiaXNzIjoiY2luZW1hdGlxdWUiLCJzdWIiOiJ1c2VyX2F1dGgiLCJleHAiOjE3NTMzNjYzMTIsImlhdCI6MTc1MzM2NTQxMn0.CpFfbOT4RoRdYXZYKi9KNmHdbCzr4gHZ99seOBz_Tco"
BASE_URL="http://localhost:8080"

# Функция для извлечения заголовков rate limit
get_rate_headers() {
    local response="$1"
    local limit=$(echo "$response" | grep -i "X-RateLimit-Limit" | cut -d' ' -f2 | tr -d '\r')
    local remaining=$(echo "$response" | grep -i "X-RateLimit-Remaining" | cut -d' ' -f2 | tr -d '\r')
    echo "Limit: $limit, Remaining: $remaining"
}

echo "=== Точное тестирование Rate Limiting ==="

# Проверяем начальное состояние
echo "1. Проверка начального состояния:"
initial_status=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
echo "Initial status: $initial_status" | jq .

# Делаем 5 запросов и отслеживаем заголовки
echo -e "\n2. Делаем 5 запросов для проверки счетчика:"
for i in {1..5}; do
    echo "Запрос $i:"
    response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/movies")
    headers=$(echo "$response" | head -10)
    echo "$headers" | grep -E "(X-RateLimit|HTTP)"
    
    # Проверяем оставшиеся запросы через отдельный вызов
    status=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
    count=$(echo "$status" | jq -r '.current_count')
    remaining=$(echo "$status" | jq -r '.remaining')
    echo "Текущий счетчик: $count, Осталось: $remaining"
done

# Проверка endpoint /api/actors
echo -e "\n3. Проверка rate limiting для /api/actors:"
actors_response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/actors")
echo "$actors_response" | head -10 | grep -E "(X-RateLimit|HTTP)"

# Проверка различных endpoint имеют ли они разные лимиты
echo -e "\n4. Проверка различных endpoint:"
endpoints=("/api/movies" "/api/actors" "/api/movies?page=1" "/api/movies?limit=10")

for endpoint in "${endpoints[@]}"; do
    echo "Проверка: $endpoint"
    response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL$endpoint")
    rate_headers=$(echo "$response" | head -15 | grep -i ratelimit)
    if [ -n "$rate_headers" ]; then
        echo "Rate limit headers: $rate_headers"
    else
        echo "No rate limit headers found"
    fi
done

echo -e "\n5. Демонстрация работы Redis (если доступен redis-cli):"
if command -v redis-cli &> /dev/null; then
    echo "Проверка Redis ключей rate limit:"
    redis-cli -p 6379 keys "*ratelimit*" 2>/dev/null || echo "Redis недоступен или ключи не найдены"
else
    echo "redis-cli не установлен, пропускаем проверку Redis"
fi

echo -e "\n=== Тест завершен ==="
