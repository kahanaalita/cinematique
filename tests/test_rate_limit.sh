#!/bin/bash

# Тест Rate Limiting для Cinematique API

set -e

API_BASE="http://localhost:8080/api"
JWT_TOKEN=""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Rate Limiting Test Suite ===${NC}"
echo

# Функция для логина и получения JWT токена
login() {
    echo -e "${YELLOW}1. Logging in to get JWT token...${NC}"
    
    RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "admin123"
        }')
    
    JWT_TOKEN=$(echo $RESPONSE | jq -r '.access_token // .token // empty')
    
    if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
        echo -e "${RED}Failed to get JWT token. Response: $RESPONSE${NC}"
        echo -e "${YELLOW}Trying to register admin user first...${NC}"
        
        # Пытаемся зарегистрировать пользователя
        curl -s -X POST "$API_BASE/auth/register" \
            -H "Content-Type: application/json" \
            -d '{
                "username": "admin",
                "password": "admin123",
                "role": "admin"
            }' > /dev/null
        
        # Пытаемся войти снова
        RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
            -H "Content-Type: application/json" \
            -d '{
                "username": "admin",
                "password": "admin123"
            }')
        
        JWT_TOKEN=$(echo $RESPONSE | jq -r '.access_token // .token // empty')
    fi
    
    if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
        echo -e "${RED}Still failed to get JWT token. Using anonymous access.${NC}"
        JWT_TOKEN=""
    else
        echo -e "${GREEN}✓ JWT token obtained${NC}"
    fi
    echo
}

# Функция для проверки статуса rate limiting
check_rate_limit_status() {
    echo -e "${YELLOW}2. Checking rate limit status...${NC}"
    
    if [ -n "$JWT_TOKEN" ]; then
        AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    else
        AUTH_HEADER="X-Anonymous: true"
    fi
    
    RESPONSE=$(curl -s -H "$AUTH_HEADER" "$API_BASE/rate-limit/status?endpoint=/api/movies")
    
    echo "Rate limit status:"
    echo $RESPONSE | jq '.' 2>/dev/null || echo $RESPONSE
    echo
}

# Функция для тестирования нормальных запросов
test_normal_requests() {
    echo -e "${YELLOW}3. Testing normal requests (should be allowed)...${NC}"
    
    if [ -n "$JWT_TOKEN" ]; then
        AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    else
        AUTH_HEADER="X-Anonymous: true"
    fi
    
    for i in {1..5}; do
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/movies")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "Request $i: ${GREEN}✓ SUCCESS (200)${NC}"
        else
            echo -e "Request $i: ${RED}✗ FAILED ($HTTP_CODE)${NC}"
            echo "Response: $(echo $RESPONSE | sed 's/HTTP_CODE:[0-9]*//')"
        fi
        
        sleep 0.1
    done
    echo
}

# Функция для тестирования превышения лимита
test_rate_limit_exceeded() {
    echo -e "${YELLOW}4. Testing rate limit exceeded (sending many requests)...${NC}"
    echo "This may take a while..."
    
    if [ -n "$JWT_TOKEN" ]; then
        AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    else
        AUTH_HEADER="X-Anonymous: true"
    fi
    
    SUCCESS_COUNT=0
    RATE_LIMITED_COUNT=0
    
    # Отправляем много запросов быстро
    for i in {1..50}; do
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE/movies")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "200" ]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            if [ $((i % 10)) -eq 0 ]; then
                echo -e "Request $i: ${GREEN}✓ SUCCESS${NC}"
            fi
        elif [ "$HTTP_CODE" = "429" ]; then
            RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
            if [ $RATE_LIMITED_COUNT -eq 1 ]; then
                echo -e "Request $i: ${RED}✗ RATE LIMITED (429)${NC}"
                echo "Rate limit response:"
                echo $RESPONSE | sed 's/HTTP_CODE:[0-9]*//' | jq '.' 2>/dev/null || echo $RESPONSE | sed 's/HTTP_CODE:[0-9]*//'
            elif [ $((RATE_LIMITED_COUNT % 10)) -eq 0 ]; then
                echo -e "Request $i: ${RED}✗ RATE LIMITED (429)${NC}"
            fi
        else
            echo -e "Request $i: ${YELLOW}? UNEXPECTED ($HTTP_CODE)${NC}"
        fi
        
        # Небольшая задержка, чтобы не перегружать систему
        sleep 0.01
    done
    
    echo
    echo -e "${BLUE}Results:${NC}"
    echo -e "  Successful requests: ${GREEN}$SUCCESS_COUNT${NC}"
    echo -e "  Rate limited requests: ${RED}$RATE_LIMITED_COUNT${NC}"
    echo
}

# Функция для проверки заголовков rate limiting
test_rate_limit_headers() {
    echo -e "${YELLOW}5. Testing rate limit headers...${NC}"
    
    if [ -n "$JWT_TOKEN" ]; then
        AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    else
        AUTH_HEADER="X-Anonymous: true"
    fi
    
    RESPONSE=$(curl -s -I -H "$AUTH_HEADER" "$API_BASE/movies")
    
    echo "Response headers:"
    echo "$RESPONSE" | grep -i "x-ratelimit" || echo "No rate limit headers found"
    echo
}

# Функция для тестирования разных endpoints
test_different_endpoints() {
    echo -e "${YELLOW}6. Testing different endpoints...${NC}"
    
    if [ -n "$JWT_TOKEN" ]; then
        AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"
    else
        AUTH_HEADER="X-Anonymous: true"
    fi
    
    ENDPOINTS=("/api/movies" "/api/actors" "/api/movies/search?title=test")
    
    for endpoint in "${ENDPOINTS[@]}"; do
        RESPONSE=$(curl -s -w "HTTP_CODE:%{http_code}" -H "$AUTH_HEADER" "$API_BASE${endpoint#/api}")
        HTTP_CODE=$(echo $RESPONSE | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "$endpoint: ${GREEN}✓ SUCCESS${NC}"
        elif [ "$HTTP_CODE" = "429" ]; then
            echo -e "$endpoint: ${RED}✗ RATE LIMITED${NC}"
        else
            echo -e "$endpoint: ${YELLOW}? STATUS $HTTP_CODE${NC}"
        fi
    done
    echo
}

# Основная функция
main() {
    echo -e "${BLUE}Starting Rate Limiting Tests...${NC}"
    echo "Make sure the application is running on localhost:8080"
    echo
    
    # Проверяем, что приложение запущено
    if ! curl -s "$API_BASE/movies" > /dev/null; then
        echo -e "${RED}Error: Application is not running on localhost:8080${NC}"
        echo "Please start the application first: go run main.go"
        exit 1
    fi
    
    login
    check_rate_limit_status
    test_normal_requests
    test_rate_limit_headers
    test_different_endpoints
    test_rate_limit_exceeded
    
    echo -e "${BLUE}=== Rate Limiting Tests Completed ===${NC}"
}

# Проверяем наличие jq
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON responses will not be formatted.${NC}"
    echo "Install jq for better output: brew install jq (macOS) or apt-get install jq (Ubuntu)"
    echo
fi

# Запускаем тесты
main