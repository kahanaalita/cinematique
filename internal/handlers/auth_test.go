package handlers

import (
	"bytes"
	"cinematique/internal/auth"
	"cinematique/internal/kafka"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of the AuthService interface
type MockAuthService struct {
	mock.Mock
}

// Ensure MockAuthService implements AuthService
var _ AuthService = (*MockAuthService)(nil)

func (m *MockAuthService) Register(username, email, password, role string) (int, error) {
	args := m.Called(username, email, password, role)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthService) Login(username, password string) (*auth.TokenPair, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

// Define error variables for testing
var (
	errUserAlreadyExists  = errors.New("user already exists")
	errInvalidCredentials = errors.New("invalid credentials")
)

func setupRouter() (*gin.Engine, *MockAuthService, *kafka.MockProducer, *AuthHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockService := new(MockAuthService)
	mockProducer := kafka.NewMockProducer()

	// Create a producer pool with the mock producer
	producerPool := kafka.NewProducerPool(mockProducer, 1, 10)
	// Create a handler with the mock service and producer pool
	handler := NewAuthHandler(mockService, producerPool)

	mockProducer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

	return r, mockService, mockProducer, handler
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService, *kafka.MockProducer)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "user",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Register", "testuser", "test@example.com", "password123", "user").Return(1, nil)
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
		{
			name: "invalid request",
			requestBody: map[string]string{
				"username": "", // Invalid: empty username
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				// Продюсер не должен вызываться при невалидном запросе
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"неверный запрос"}`,
		},
		{
			name: "registration error",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "user",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Register", "testuser", "test@example.com", "password123", "user").Return(0, errUserAlreadyExists)
				// Продюсер не должен вызываться при ошибке регистрации
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"user already exists"}`,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				// Missing password
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				// Продюсер не должен вызываться при невалидном запросе
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"неверный запрос"}`,
		},
		{
			name: "produce error",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
				"role":     "user",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Register", "testuser", "test@example.com", "password123", "user").Return(1, nil)
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(errors.New("kafka produce error"))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, mockProducer, handler := setupRouter()
			tt.setupMock(mockService, mockProducer)

			r.POST("/register", handler.Register)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			mockService.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService, *kafka.MockProducer)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Login", "testuser", "password123").Return(&auth.TokenPair{
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
					ExpiresIn:    3600,
				}, nil)
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"access_token":"test_access_token","refresh_token":"test_refresh_token","expires_in":3600}`,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
				// Missing password
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				// Продюсер не должен вызываться при невалидном запросе
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"неверный запрос"}`,
		},
		{
			name: "service error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Login", "testuser", "password123").Return((*auth.TokenPair)(nil), errors.New("internal server error"))
				// Продюсер не должен вызываться при ошибке
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Login", "testuser", "wrongpassword").Return((*auth.TokenPair)(nil), errInvalidCredentials)
				// Продюсер не должен вызываться при неверных учетных данных
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
		{
			name: "produce error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("Login", "testuser", "password123").Return(&auth.TokenPair{
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
					ExpiresIn:    3600,
				}, nil)
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(errors.New("kafka produce error"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"access_token":"test_access_token","refresh_token":"test_refresh_token","expires_in":3600}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, mockProducer, handler := setupRouter()
			tt.setupMock(mockService, mockProducer)

			r.POST("/login", handler.Login)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			mockService.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService, *kafka.MockProducer)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"refresh_token": "valid_refresh_token",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("RefreshToken", "valid_refresh_token").Return(&auth.TokenPair{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					ExpiresIn:    3600,
				}, nil)
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"access_token":"new_access_token","refresh_token":"new_refresh_token","expires_in":3600}`,
		},
		{
			name: "invalid token",
			requestBody: map[string]string{
				"refresh_token": "invalid_token",
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				m.On("RefreshToken", "invalid_token").Return((*auth.TokenPair)(nil), errors.New("invalid refresh token"))
				p.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid refresh token"}`,
		},
		{
			name:        "missing token",
			requestBody: map[string]string{},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				// Продюсер не должен вызываться при невалидном запросе
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"неверный запрос"}`,
		},
		{
			name: "invalid request",
			requestBody: map[string]string{
				"refresh_token": "", // Invalid: empty token
			},
			setupMock: func(m *MockAuthService, p *kafka.MockProducer) {
				// Продюсер не должен вызываться при невалидном запросе
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"неверный запрос"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, mockProducer, handler := setupRouter()
			r.POST("/refresh", handler.Refresh)

			// Настраиваем моки
			tt.setupMock(mockService, mockProducer)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Проверяем HTTP ответ
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			// Проверяем, что все ожидаемые вызовы были сделаны
			mockService.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		expectBody     bool
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"refresh_token": "valid_refresh_token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Logout", "valid_refresh_token").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectBody:     false,
		},
		{
			name: "invalid token",
			requestBody: map[string]string{
				"refresh_token": "invalid_token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Logout", "invalid_token").Return(errors.New("failed to logout"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectBody:     true,
			expectedBody:   `{"error":"failed to logout"}`,
		},
		{
			name:        "missing token",
			requestBody: map[string]string{},
			setupMock: func(m *MockAuthService) {
				// Продюсер не должен вызываться при невалидном запросе
			},
			expectedStatus: http.StatusBadRequest,
			expectBody:     true,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "invalid request",
			requestBody: map[string]string{
				"refresh_token": "", // Invalid: empty token
			},
			setupMock: func(m *MockAuthService) {
				// Продюсер не должен вызываться при невалидном запросе
			},
			expectedStatus: http.StatusBadRequest,
			expectBody:     true,
			expectedBody:   `{"error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, _, handler := setupRouter()
			r.POST("/logout", handler.Logout)

			tt.setupMock(mockService)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/logout", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectBody {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
