#!/bin/bash

# Тестирование Keycloak интеграции
echo "🚀 Testing Keycloak Integration"
echo "================================"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# URL приложения
BASE_URL="http://localhost:8080/api"

# Функция для тестирования эндпоинта
test_endpoint() {
    local method=$1
    local endpoint=$2
    local token=$3
    local description=$4
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Method: $method"
    echo "Endpoint: $endpoint"
    echo "Token type: $(echo $token | cut -c1-20)..."
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -X $method \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d '{"title":"Test Movie","description":"Test Description","release_date":"2024-01-01","rating":8.5}' \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ Success (HTTP $http_code)${NC}"
        echo "Response: $(echo $body | jq . 2>/dev/null || echo $body | head -c 100)..."
    else
        echo -e "${RED}❌ Failed (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
}

# 1. Сначала получаем JWT токен через регистрацию и логин
echo -e "\n${YELLOW}Step 1: Getting JWT token${NC}"

# Регистрация пользователя
echo "Registering user..."
register_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","email":"test@example.com","password":"testpass123","role":"user"}' \
    "$BASE_URL/auth/register")

register_code=$(echo "$register_response" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$register_code" = "201" ]; then
    echo -e "${GREEN}✅ User registered successfully${NC}"
else
    echo -e "${YELLOW}⚠️  User might already exist (HTTP $register_code)${NC}"
fi

# Логин для получения JWT токена
echo "Logging in to get JWT token..."
login_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"testpass123"}' \
    "$BASE_URL/auth/login")

login_code=$(echo "$login_response" | grep "HTTP_CODE:" | cut -d: -f2)
login_body=$(echo "$login_response" | sed '/HTTP_CODE:/d')

if [ "$login_code" = "200" ]; then
    JWT_TOKEN=$(echo $login_body | jq -r '.access_token' 2>/dev/null)
    echo -e "${GREEN}✅ JWT token obtained${NC}"
    echo "JWT Token: ${JWT_TOKEN:0:50}..."
else
    echo -e "${RED}❌ Failed to get JWT token (HTTP $login_code)${NC}"
    echo "Response: $login_body"
    exit 1
fi

# 2. Создаем mock Keycloak токен для тестирования
echo -e "\n${YELLOW}Step 2: Creating mock Keycloak token${NC}"

# Создаем простой mock токен с правильной структурой Keycloak
# В реальности это будет токен от Keycloak сервера
KEYCLOAK_MOCK_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJyc2EtZ2VuZXJhdGVkIn0.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwianRpIjoiYWJjZGVmZ2gtaWprbC1tbm9wLXFyc3QtdXZ3eHl6MTIzNCIsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MC9yZWFsbXMvY2luZW1hdGlxdWUiLCJhdWQiOiJjaW5lbWF0aXF1ZS1hcGkiLCJzdWIiOiJrZXljbG9hay11c2VyLTEyMyIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNpbmVtYXRpcXVlLWFwaSIsInNlc3Npb25fc3RhdGUiOiJzZXNzaW9uLTEyMyIsImFjciI6IjEiLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7ImNpbmVtYXRpcXVlLWFwaSI6eyJyb2xlcyI6WyJjaW5lbWF0aXF1ZS11c2VyIl19fSwic2NvcGUiOiJvcGVuaWQgZW1haWwgcHJvZmlsZSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJuYW1lIjoiS2V5Y2xvYWsgVGVzdCBVc2VyIiwicHJlZmVycmVkX3VzZXJuYW1lIjoia2V5Y2xvYWt1c2VyIiwiZ2l2ZW5fbmFtZSI6IktleWNsb2FrIiwiZmFtaWx5X25hbWUiOiJVc2VyIiwiZW1haWwiOiJrZXljbG9ha0BleGFtcGxlLmNvbSJ9.mock-signature"

echo "Mock Keycloak Token: ${KEYCLOAK_MOCK_TOKEN:0:50}..."

# 3. Тестируем эндпоинты с JWT токеном
echo -e "\n${YELLOW}Step 3: Testing endpoints with JWT token${NC}"

test_endpoint "GET" "/movies" "$JWT_TOKEN" "Get movies with JWT token"
test_endpoint "GET" "/actors" "$JWT_TOKEN" "Get actors with JWT token"

# 4. Тестируем эндпоинты с Keycloak токеном (если включен)
echo -e "\n${YELLOW}Step 4: Testing endpoints with Keycloak token${NC}"

# Проверяем, включен ли Keycloak
if grep -q "KEYCLOAK_ENABLED=true" .env; then
    echo "Keycloak is enabled, testing with Keycloak token..."
    test_endpoint "GET" "/movies" "$KEYCLOAK_MOCK_TOKEN" "Get movies with Keycloak token"
    test_endpoint "GET" "/actors" "$KEYCLOAK_MOCK_TOKEN" "Get actors with Keycloak token"
else
    echo -e "${YELLOW}⚠️  Keycloak is disabled in .env file${NC}"
    echo "To test Keycloak tokens, set KEYCLOAK_ENABLED=true in .env"
fi

# 5. Тестируем неавторизованные запросы
echo -e "\n${YELLOW}Step 5: Testing unauthorized requests${NC}"

echo "Testing without token..."
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$BASE_URL/movies")
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✅ Correctly rejected unauthorized request (HTTP $http_code)${NC}"
else
    echo -e "${RED}❌ Should have rejected unauthorized request (HTTP $http_code)${NC}"
fi

# 6. Тестируем недействительные токены
echo -e "\n${YELLOW}Step 6: Testing invalid tokens${NC}"

test_endpoint "GET" "/movies" "invalid-token" "Invalid token test"

echo -e "\n${GREEN}🎉 Testing completed!${NC}"
echo "================================"