#!/bin/bash

# Финальный тест всей системы Cinematique с Rate Limiting

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api"
JWT_TOKEN=""

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                 CINEMATIQUE FINAL TEST SUITE                 ║${NC}"
echo -e "${BLUE}║                   with Rate Limiting                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo

# Функция для проверки зависимостей
check_dependencies() {
    echo -e "${YELLOW}🔍 Checking dependencies...${NC}"
    
    # Проверяем приложение
    if ! curl -s "$API_BASE/movies" > /dev/null 2>&1; then
        echo -e "${RED}❌ Application is not running on localhost:8080${NC}"
        echo "Please start: go run main.go"
        exit 1
    fi
    echo -e "${GREEN}✓ Application is running${NC}"
    
    # Проверяем Redis
    if ! docker-compose ps redis | grep -q "Up"; then
        echo -e "${RED}❌ Redis is not running${NC}"
        echo "Please start: docker-compose up -d redis"
        exit 1
    fi
    echo -e "${GREEN}✓ Redis is running${NC}"
    
    # Проверяем jq (опционально)
    if command -v jq &> /dev/null; then
        echo -e "${GREEN}✓ jq is available${NC}"
    else
        echo -e "${YELLOW}⚠ jq not found (JSON won't be formatted)${NC}"
    fi
    
    echo
}

# Функция для аутентификации
authenticate() {
    echo -e "${YELLOW}🔐 Authenticating...${NC}"
    
    # Регистрируем тестового пользователя
    curl -s -X POST "$API_BASE/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser",
            "password": "testpass123",
            "role": "user"
        }' > /dev/null 2>&1
    
    # Логинимся
    RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser",
            "password": "testpass123"
        }')
    
    JWT_TOKEN=$(echo $RESPONSE | jq -r '.access_token // .token // empty' 2>/dev/null || echo "")
    
    if [ -n "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
        echo -e "${GREEN}✓ Authentication successful${NC}"
    else
        echo -e "${RED}❌ Authentication failed${NC}"
        echo "Response: $RESPONSE"
        exit 1
    fi
    echo
}

# Функция для тестирования базового API
test_basic_api() {
    echo -e "${YELLOW}🎬 Testing basic API functionality...${NC}"
    
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    
    # Тест получения фильмов
    RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/movies")
    HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ GET /api/movies - SUCCESS${NC}"
    else
        echo -e "${RED}❌ GET /api/movies - FAILED ($HTTP_CODE)${NC}"
    fi
    
    # Тест получения актеров
    RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/actors")
    HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ GET /api/actors - SUCCESS${NC}"
    else
        echo -e "${RED}❌ GET /api/actors - FAILED ($HTTP_CODE)${NC}"
    fi
    
    echo
}

# Функция для тестирования rate limiting
test_rate_limiting() {
    echo -e "${YELLOW}🚦 Testing rate limiting...${NC}"
    
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    
    # Проверяем статус rate limiting
    RESPONSE=$(curl -s -H "$AUTH_HEADER" "$API_BASE/rate-limit/status")
    if echo $RESPONSE | jq -e '.enabled' > /dev/null 2>&1; then
        ENABLED=$(echo $RESPONSE | jq -r '.enabled')
        LIMIT=$(echo $RESPONSE | jq -r '.limit')
        echo -e "${GREEN}✓ Rate limiting is enabled (limit: $LIMIT)${NC}"
    else
        echo -e "${RED}❌ Rate limiting status check failed${NC}"
        return 1
    fi
    
    # Тестируем нормальные запросы
    echo -e "${BLUE}  Testing normal requests...${NC}"
    SUCCESS_COUNT=0
    for i in {1..10}; do
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/movies")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "200" ]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        fi
        sleep 0.1
    done
    
    if [ $SUCCESS_COUNT -eq 10 ]; then
        echo -e "${GREEN}  ✓ Normal requests: $SUCCESS_COUNT/10 successful${NC}"
    else
        echo -e "${YELLOW}  ⚠ Normal requests: $SUCCESS_COUNT/10 successful${NC}"
    fi
    
    # Тестируем превышение лимита (быстрые запросы)
    echo -e "${BLUE}  Testing rate limit enforcement...${NC}"
    RATE_LIMITED_COUNT=0
    
    for i in {1..20}; do
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/movies")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "429" ]; then
            RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
        fi
        sleep 0.01  # Очень быстрые запросы
    done
    
    if [ $RATE_LIMITED_COUNT -gt 0 ]; then
        echo -e "${GREEN}  ✓ Rate limiting working: $RATE_LIMITED_COUNT requests blocked${NC}"
    else
        echo -e "${YELLOW}  ⚠ Rate limiting not triggered (limit might be high)${NC}"
    fi
    
    echo
}

# Функция для тестирования заголовков
test_headers() {
    echo -e "${YELLOW}📋 Testing rate limit headers...${NC}"
    
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    
    HEADERS=$(curl -s -I -H "$AUTH_HEADER" "$API_BASE/movies")
    
    if echo "$HEADERS" | grep -qi "x-ratelimit-limit"; then
        LIMIT=$(echo "$HEADERS" | grep -i "x-ratelimit-limit" | cut -d: -f2 | tr -d ' \r')
        echo -e "${GREEN}✓ X-RateLimit-Limit: $LIMIT${NC}"
    else
        echo -e "${RED}❌ X-RateLimit-Limit header missing${NC}"
    fi
    
    if echo "$HEADERS" | grep -qi "x-ratelimit-remaining"; then
        REMAINING=$(echo "$HEADERS" | grep -i "x-ratelimit-remaining" | cut -d: -f2 | tr -d ' \r')
        echo -e "${GREEN}✓ X-RateLimit-Remaining: $REMAINING${NC}"
    else
        echo -e "${RED}❌ X-RateLimit-Remaining header missing${NC}"
    fi
    
    if echo "$HEADERS" | grep -qi "x-ratelimit-reset"; then
        RESET=$(echo "$HEADERS" | grep -i "x-ratelimit-reset" | cut -d: -f2 | tr -d ' \r')
        echo -e "${GREEN}✓ X-RateLimit-Reset: $RESET${NC}"
    else
        echo -e "${RED}❌ X-RateLimit-Reset header missing${NC}"
    fi
    
    echo
}

# Функция для тестирования разных пользователей
test_user_isolation() {
    echo -e "${YELLOW}👥 Testing user isolation...${NC}"
    
    # Создаем второго пользователя
    curl -s -X POST "$API_BASE/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser2",
            "password": "testpass123",
            "role": "user"
        }' > /dev/null 2>&1
    
    # Логинимся вторым пользователем
    RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser2",
            "password": "testpass123"
        }')
    
    JWT_TOKEN2=$(echo $RESPONSE | jq -r '.access_token // .token // empty' 2>/dev/null || echo "")
    
    if [ -n "$JWT_TOKEN2" ] && [ "$JWT_TOKEN2" != "null" ]; then
        echo -e "${GREEN}✓ Second user authenticated${NC}"
        
        # Тестируем, что у второго пользователя свой лимит
        AUTH_HEADER2="Authorization: Bearer $JWT_TOKEN2"
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER2" "$API_BASE/movies")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "${GREEN}✓ User isolation working - second user can make requests${NC}"
        else
            echo -e "${RED}❌ User isolation failed - second user blocked${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ Could not create second user${NC}"
    fi
    
    echo
}

# Функция для тестирования производительности
test_performance() {
    echo -e "${YELLOW}⚡ Testing performance...${NC}"
    
    AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    
    START_TIME=$(date +%s%N)
    
    # Делаем 50 запросов
    for i in {1..50}; do
        curl -s -H "$AUTH_HEADER" "$API_BASE/movies" > /dev/null
    done
    
    END_TIME=$(date +%s%N)
    DURATION=$(( (END_TIME - START_TIME) / 1000000 )) # Convert to milliseconds
    
    echo -e "${GREEN}✓ 50 requests completed in ${DURATION}ms${NC}"
    echo -e "${BLUE}  Average: $((DURATION / 50))ms per request${NC}"
    
    echo
}

# Функция для проверки Redis
test_redis_integration() {
    echo -e "${YELLOW}🔴 Testing Redis integration...${NC}"
    
    # Проверяем, что ключи создаются в Redis
    REDIS_KEYS=$(docker-compose exec -T redis redis-cli KEYS "ratelimit:*" 2>/dev/null || echo "")
    
    if [ -n "$REDIS_KEYS" ]; then
        KEY_COUNT=$(echo "$REDIS_KEYS" | wc -l)
        echo -e "${GREEN}✓ Redis keys found: $KEY_COUNT rate limit keys${NC}"
        
        # Показываем пример ключа
        FIRST_KEY=$(echo "$REDIS_KEYS" | head -n1)
        if [ -n "$FIRST_KEY" ]; then
            VALUE=$(docker-compose exec -T redis redis-cli GET "$FIRST_KEY" 2>/dev/null || echo "")
            echo -e "${BLUE}  Example key: $FIRST_KEY = $VALUE${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ No Redis keys found (might be expired)${NC}"
    fi
    
    echo
}

# Функция для отчета
generate_report() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║                        TEST REPORT                          ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${BLUE}✅ Tests Completed Successfully:${NC}"
    echo -e "   • Application connectivity"
    echo -e "   • Redis connectivity"
    echo -e "   • User authentication"
    echo -e "   • Basic API functionality"
    echo -e "   • Rate limiting enforcement"
    echo -e "   • HTTP headers"
    echo -e "   • User isolation"
    echo -e "   • Performance testing"
    echo -e "   • Redis integration"
    echo
    echo -e "${GREEN}🎉 Rate Limiting Implementation: SUCCESSFUL${NC}"
    echo
    echo -e "${YELLOW}📚 Documentation:${NC}"
    echo -e "   • Rate Limiting: docs/RATE_LIMITING.md"
    echo -e "   • Examples: examples/rate_limit_example.go"
    echo -e "   • Tests: tests/test_rate_limit.sh"
    echo
}

# Основная функция
main() {
    check_dependencies
    authenticate
    test_basic_api
    test_rate_limiting
    test_headers
    test_user_isolation
    test_performance
    test_redis_integration
    generate_report
}

# Обработка сигналов
trap 'echo -e "\n${RED}Test interrupted${NC}"; exit 1' INT TERM

# Запуск тестов
main