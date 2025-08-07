#!/bin/bash

# Rate Limit тестирование для Cinematique API

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGUiOiJ1c2VyIiwiaXNzIjoiY2luZW1hdGlxdWUiLCJzdWIiOiJ1c2VyX2F1dGgiLCJleHAiOjE3NTMzNjYzMTIsImlhdCI6MTc1MzM2NTQxMn0.CpFfbOT4RoRdYXZYKi9KNmHdbCzr4gHZ99seOBz_Tco"
BASE_URL="http://localhost:8080"

# Функция для вывода заголовков ответа
print_headers() {
    echo "=== Response Headers ==="
    echo "$1" | grep -E "(X-RateLimit|HTTP|429|200|403)"
    echo "========================"
}

echo "=== Тест 1: Проверка текущего статуса rate limit ==="
curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies" | jq .

echo -e "\n=== Тест 2: Последовательные запросы для проверки rate limit ==="
echo "Делаем 10 запросов к /api/movies для демонстрации работы счетчика..."

for i in {1..10}; do
    echo "Запрос $i..."
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}\nHEADERS:%{header_json}\n" -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/movies")
    echo "$response" | grep -E "(HTTP_CODE|X-RateLimit)"
done

echo -e "\n=== Тест 3: Проверка rate limit headers в обычных запросах ==="
response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/movies")
echo "$response" | head -20

echo -e "\n=== Тест 4: Проверка защищенных endpoint ==="
endpoints=("/api/movies" "/api/actors" "/api/movies/1")

for endpoint in "${endpoints[@]}"; do
    echo "Проверка $endpoint..."
    curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL$endpoint" -o /dev/null -w "Endpoint: $endpoint - Status: %{http_code}, RateLimit-Remaining: %{header_json}\\n"
done

echo -e "\n=== Тест 5: Проверка 429 ошибки (без реального превышения лимита) ==="
echo "Для проверки 429 ошибки можно использовать следующий цикл:"
echo "for i in {1..1005}; do curl -H \"Authorization: Bearer $TOKEN\" $BASE_URL/api/movies; done"
echo "Но в данный момент пропустим этот тест, чтобы не нагружать систему"

echo -e "\n=== Тест завершен ==="
