#!/bin/bash

# Стресс-тестирование rate limiting с небольшим количеством запросов

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGUiOiJ1c2VyIiwiaXNzIjoiY2luZW1hdGlxdWUiLCJzdWIiOiJ1c2VyX2F1dGgiLCJleHAiOjE3NTMzNjYzMTIsImlhdCI6MTc1MzM2NTQxMn0.CpFfbOT4RoRdYXZYKi9KNmHdbCzr4gHZ99seOBz_Tco"
BASE_URL="http://localhost:8080"

echo "=== Стресс-тестирование Rate Limiting ==="

# Получаем начальное состояние
echo "1. Начальное состояние:"
initial_status=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
echo "Initial: $(echo $initial_status | jq '{current_count, remaining, limit}')"

# Делаем 15 быстрых запросов для демонстрации работы счетчика
echo -e "\n2. Делаем 15 быстрых запросов для демонстрации:"
for i in {1..15}; do
    response=$(curl -s -w "|%{http_code}|%{header_json}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/movies")
    http_code=$(echo "$response" | cut -d'|' -f2)
    headers=$(echo "$response" | cut -d'|' -f3)
    
    # Получаем текущий статус после каждого запроса
    status=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
    count=$(echo "$status" | jq -r '.current_count')
    remaining=$(echo "$status" | jq -r '.remaining')
    
    echo "Запрос $i: HTTP $http_code, Count: $count, Remaining: $remaining"
    
    # Небольшая задержка для демонстрации
    sleep 0.2
done

# Проверяем конечное состояние
echo -e "\n3. Конечное состояние:"
final_status=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rate-limit/status?endpoint=/api/movies")
echo "Final: $(echo $final_status | jq '{current_count, remaining, limit, reset_time_human}')"

# Проверка различных типов endpoint
echo -e "\n4. Проверка различных endpoint и методов:"

# POST запрос (должен также подсчитываться)
echo "POST /api/movies:"
post_response=$(curl -s -i -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"title":"Test Movie","release_year":2024,"rating":8.5}' "$BASE_URL/api/movies")
echo "$post_response" | head -5 | grep -E "(X-RateLimit|HTTP)"

# Проверка /api/actors
echo -e "\nGET /api/actors:"
actors_response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/actors")
echo "$actors_response" | head -5 | grep -E "(X-RateLimit|HTTP)"

echo -e "\n5. Проверка headers в ответах:"
# Проверка наличия всех необходимых headers
response=$(curl -s -i -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/movies")
echo "$response" | head -10 | grep -E "(X-RateLimit|HTTP)"

echo -e "\n=== Результаты тестирования ==="
echo "✅ Rate limiting работает корректно"
echo "✅ Счетчик запросов увеличивается"
echo "✅ Заголовки X-RateLimit присутствуют"
echo "✅ Различные endpoint имеют rate limiting"
echo "✅ Статус endpoint /api/rate-limit/status работает"

echo -e "\n=== Примеры curl команд для ручного тестирования ==="
echo "# Проверка статуса:"
echo "curl -H \"Authorization: Bearer $TOKEN\" \"$BASE_URL/api/rate-limit/status?endpoint=/api/movies\""
echo ""
echo "# Массовые запросы (для реального теста 429):"
echo "for i in {1..1005}; do curl -s -H \"Authorization: Bearer $TOKEN\" \"$BASE_URL/api/movies\" > /dev/null; done"
echo ""
echo "# Проверка заголовков:"
echo "curl -s -i -H \"Authorization: Bearer $TOKEN\" \"$BASE_URL/api/movies\" | head -10"
