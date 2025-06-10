package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockAuthService is a mock implementation of the AuthService interface
type MockAuthService struct {
	mock.Mock
}

// Ensure MockAuthService implements AuthService
var _ AuthService = (*MockAuthService)(nil)

func (m *MockAuthService) Register(username, password, role string) (int, error) {
	args := m.Called(username, password, role)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

// Define error variables for testing
var (
	errUserAlreadyExists = errors.New("user already exists")
	errInvalidCredentials = errors.New("invalid credentials")
)

func setupRouter() (*gin.Engine, *MockAuthService, *AuthHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockService := new(MockAuthService)
	// Create a handler with the mock service
	handler := &AuthHandler{service: mockService}
	return r, mockService, handler
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock     func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
				"role":     "user",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Register", "testuser", "password123", "user").Return(1, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
		{
			name: "invalid request",
			requestBody: map[string]string{
				"username": "", // Invalid: empty username
				"password": "password123",
			},
			setupMock:     func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "registration error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
				"role":     "user",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Register", "testuser", "password123", "user").Return(0, errUserAlreadyExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"user already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, handler := setupRouter()
			tt.setupMock(mockService)

			r.POST("/register", handler.Register)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock     func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", "testuser", "password123").Return("test.token.123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"test.token.123"}`,
		},
		{
			name: "invalid request",
			requestBody: map[string]string{
				"username": "testuser",
				// Missing password
			},
			setupMock:     func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", "testuser", "wrongpassword").Return("", errInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockService, handler := setupRouter()
			tt.setupMock(mockService)

			r.POST("/login", handler.Login)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
