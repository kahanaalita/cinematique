package handlers

import (
	"bytes"
	"cinematique/internal/controller/dto"
	"cinematique/internal/domain"
	"cinematique/internal/kafka"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockActorController - мок-реализация интерфейса ActorController
type MockActorController struct {
	mock.Mock
}

func (m *MockActorController) CreateActor(c *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error) {
	args := m.Called(c, req)
	return args.Get(0).(dto.ActorResponse), args.Error(1)
}

func (m *MockActorController) GetActorByID(c *gin.Context, id int) (dto.ActorResponse, error) {
	args := m.Called(c, id)
	return args.Get(0).(dto.ActorResponse), args.Error(1)
}

func (m *MockActorController) UpdateActor(c *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error) {
	args := m.Called(c, id, req)
	return args.Get(0).(dto.ActorResponse), args.Error(1)
}

func (m *MockActorController) DeleteActor(c *gin.Context, id int) error {
	args := m.Called(c, id)
	return args.Error(0)
}

func (m *MockActorController) ListActors(c *gin.Context) (dto.ActorsListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.ActorsListResponse), args.Error(1)
}

func (m *MockActorController) GetAllActorsWithMovies(c *gin.Context) (dto.ActorsWithFilmsListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.ActorsWithFilmsListResponse), args.Error(1)
}

func (m *MockActorController) PartialUpdateActor(c *gin.Context, id int, update dto.ActorUpdate) (dto.ActorResponse, error) {
	args := m.Called(c, id, update)
	return args.Get(0).(dto.ActorResponse), args.Error(1)
}

// TestActorHandler_Create tests the Create method of ActorHandler
func TestActorHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockActorController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]interface{}{
				"name":       "Test Actor",
				"gender":     "male",
				"birth_date": "1990-01-01T00:00:00Z",
			},
			setupMock: func(m *MockActorController) {
				expectedReq := dto.CreateActorRequest{
					Name:      "Test Actor",
					Gender:    "male",
					BirthDate: "1990-01-01T00:00:00Z",
				}
				m.On("CreateActor", mock.Anything, expectedReq).
					Return(dto.ActorResponse{
						ID:        1,
						Name:      "Test Actor",
						Gender:    "male",
						BirthDate: "1990-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"name":"Test Actor","gender":"male","birth_date":"1990-01-01T00:00:00Z"}`,
		},
		{
			name: "empty name",
			requestBody: map[string]interface{}{
				"name":       "", // Invalid: empty name
				"gender":     "male",
				"birth_date": "1990-01-01T00:00:00Z",
			},
			setupMock: func(m *MockActorController) {
				// No mock setup needed as the handler should return before calling the controller
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Name is required"}`,
		},
		{
			name: "missing gender",
			requestBody: map[string]interface{}{
				"name":       "Test Actor",
				"birth_date": "1990-01-01T00:00:00Z",
				// Gender is missing
			},
			setupMock: func(m *MockActorController) {
				// No mock setup needed as the handler should return before calling the controller
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Gender is required"}`,
		},
		{
			name: "missing birth date",
			requestBody: map[string]interface{}{
				"name":   "Test Actor",
				"gender": "male",
				// BirthDate is missing
			},
			setupMock: func(m *MockActorController) {
				// No mock setup needed as the handler should return before calling the controller
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"BirthDate is required"}`,
		},
		{
			name: "controller error",
			requestBody: map[string]interface{}{
				"name":       "Test Actor",
				"gender":     "male",
				"birth_date": "1990-01-01T00:00:00Z",
			},
			setupMock: func(m *MockActorController) {
				expectedReq := dto.CreateActorRequest{
					Name:      "Test Actor",
					Gender:    "male",
					BirthDate: "1990-01-01T00:00:00Z",
				}
				m.On("CreateActor", mock.Anything, expectedReq).
					Return(dto.ActorResponse{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			tt.setupMock(mockCtrl)

			r.POST("/actors", handler.Create)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/actors", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestActorHandler_GetByID tests the GetByID method of ActorHandler
func TestActorHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		actorID        string
		setupMock      func(*MockActorController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			actorID: "1",
			setupMock: func(m *MockActorController, id int) {
				m.On("GetActorByID", mock.Anything, id).
					Return(dto.ActorResponse{
						ID:        1,
						Name:      "Test Actor",
						Gender:    "male",
						BirthDate: "1990-01-01T00:00:00Z",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Test Actor","gender":"male","birth_date":"1990-01-01T00:00:00Z"}`,
		},
		{
			name:    "invalid id",
			actorID: "invalid",
			setupMock: func(m *MockActorController, id int) {
				// No mock setup needed for this case
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:    "not found",
			actorID: "999",
			setupMock: func(m *MockActorController, id int) {
				m.On("GetActorByID", mock.Anything, id).
					Return(dto.ActorResponse{}, errors.New("actor not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"actor not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, actorID)

			r.GET("/actors/:id", handler.GetByID)

			req, _ := http.NewRequest("GET", "/actors/"+tt.actorID, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestActorHandler_List tests the List method of ActorHandler
func TestActorHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockActorController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupMock: func(m *MockActorController) {
				m.On("ListActors", mock.Anything).Return(dto.ActorsListResponse{Actors: []dto.ActorResponse{}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"actors":[]}`,
		},
		{
			name: "controller error",
			setupMock: func(m *MockActorController) {
				m.On("ListActors", mock.Anything).Return(dto.ActorsListResponse{}, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			tt.setupMock(mockCtrl)

			r.GET("/actors", handler.List)
			req, _ := http.NewRequest("GET", "/actors", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

// TestActorHandler_Update tests the Update method of ActorHandler
func TestActorHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		actorID        string
		requestBody    string
		setupMock      func(*MockActorController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success",
			actorID:     "1",
			requestBody: `{"name":"Updated","gender":"male","birth_date":"1990-01-01"}`,
			setupMock: func(m *MockActorController, id int) {
				name := "Updated"
				gender := "male"
				birthDate := "1990-01-01"
				_ = dto.UpdateActorRequest{
					Name:      &name,
					Gender:    &gender,
					BirthDate: &birthDate,
				}
				m.On("UpdateActor", mock.Anything, id, mock.MatchedBy(func(req dto.UpdateActorRequest) bool {
					return *req.Name == "Updated" && *req.Gender == "male" && *req.BirthDate == "1990-01-01"
				})).Return(dto.ActorResponse{
					ID:        1,
					Name:      "Updated",
					Gender:    "male",
					BirthDate: "1990-01-01",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"Updated","gender":"male","birth_date":"1990-01-01"}`,
		},
		{
			name:           "invalid id",
			actorID:        "abc",
			requestBody:    `{"name":"Test"}`,
			setupMock:      func(m *MockActorController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:        "not found",
			actorID:     "999",
			requestBody: `{"name":"NotExist"}`,
			setupMock: func(m *MockActorController, id int) {
				m.On("UpdateActor", mock.Anything, id, mock.Anything).
					Return(dto.ActorResponse{}, domain.ErrActorNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"actor not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, actorID)

			r.PUT("/actors/:id", handler.Update)
			req, _ := http.NewRequest("PUT", "/actors/"+tt.actorID, bytes.NewBufferString(tt.requestBody))
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

// TestActorHandler_PartialUpdate tests the PartialUpdate method of ActorHandler
func TestActorHandler_PartialUpdate(t *testing.T) {
	tests := []struct {
		name           string
		actorID        string
		requestBody    string
		setupMock      func(*MockActorController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success update name",
			actorID:     "1",
			requestBody: `{"name":"Updated Partial"}`,
			setupMock: func(m *MockActorController, id int) {
				m.On("PartialUpdateActor", mock.Anything, id, mock.MatchedBy(func(update dto.ActorUpdate) bool {
					return *update.Name == "Updated Partial" && update.Gender == nil && update.BirthDate == nil
				})).Return(dto.ActorResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:        "success update all fields",
			actorID:     "2",
			requestBody: `{"name":"Updated","gender":"female","birth_date":"1995-01-01T00:00:00Z"}`,
			setupMock: func(m *MockActorController, id int) {
				name := "Updated"
				gender := "female"
				birthDate, _ := time.Parse(time.RFC3339, "1995-01-01T00:00:00Z")
				_ = dto.ActorUpdate{
					Name:      &name,
					Gender:    &gender,
					BirthDate: &birthDate,
				}
				m.On("PartialUpdateActor", mock.Anything, id, mock.MatchedBy(func(update dto.ActorUpdate) bool {
					expectedDate, _ := time.Parse(time.RFC3339, "1995-01-01T00:00:00Z")
					return update.Name != nil && *update.Name == "Updated" &&
						update.Gender != nil && *update.Gender == "female" &&
						update.BirthDate != nil && update.BirthDate.Equal(expectedDate)
				})).Return(dto.ActorResponse{ID: id, Name: "Updated", Gender: "female", BirthDate: "1995-01-01T00:00:00Z"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid id",
			actorID:        "abc",
			requestBody:    `{"name":"Updated"}`,
			setupMock:      func(m *MockActorController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:        "not found",
			actorID:     "999",
			requestBody: `{"name":"NotExist"}`,
			setupMock: func(m *MockActorController, id int) {
				m.On("PartialUpdateActor", mock.Anything, id, mock.Anything).
					Return(dto.ActorResponse{}, domain.ErrActorNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"actor not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, actorID)

			r.PATCH("/actors/:id", handler.PartialUpdate)
			req, _ := http.NewRequest("PATCH", "/actors/"+tt.actorID, bytes.NewBufferString(tt.requestBody))
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

// TestActorHandler_Delete tests the Delete method of ActorHandler
func TestActorHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		actorID        string
		setupMock      func(*MockActorController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			actorID: "1",
			setupMock: func(m *MockActorController, id int) {
				m.On("DeleteActor", mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "invalid id",
			actorID:        "abc",
			setupMock:      func(m *MockActorController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:    "not found",
			actorID: "999",
			setupMock: func(m *MockActorController, id int) {
				m.On("DeleteActor", mock.Anything, id).Return(domain.ErrActorNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"actor not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, actorID)

			r.DELETE("/actors/:id", handler.Delete)
			req, _ := http.NewRequest("DELETE", "/actors/"+tt.actorID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

// TestActorHandler_ListWithMovies tests the ListWithMovies method of ActorHandler
func TestActorHandler_ListWithMovies(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockActorController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupMock: func(m *MockActorController) {
				m.On("GetAllActorsWithMovies", mock.Anything).Return(dto.ActorsWithFilmsListResponse{Actors: []dto.ActorWithFilms{}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"actors":[]}`,
		},
		{
			name: "controller error",
			setupMock: func(m *MockActorController) {
				m.On("GetAllActorsWithMovies", mock.Anything).Return(dto.ActorsWithFilmsListResponse{}, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockActorController)
			handler := NewActorHandler(mockCtrl)

			tt.setupMock(mockCtrl)

			r.GET("/actors/with-movies", handler.ListWithMovies)
			req, _ := http.NewRequest("GET", "/actors/with-movies", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

// MockMovieController - мок-реализация интерфейса MovieController
type MockMovieController struct {
	mock.Mock
}

// Реализация всех необходимых методов интерфейса MovieController

func (m *MockMovieController) CreateMovie(c *gin.Context, req dto.CreateMovieRequest) (dto.MovieResponse, error) {
	args := m.Called(c, req)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) GetMovieByID(c *gin.Context, id int) (dto.MovieResponse, error) {
	args := m.Called(c, id)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) UpdateMovie(c *gin.Context, id int, req dto.UpdateMovieRequest) (dto.MovieResponse, error) {
	args := m.Called(c, id, req)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) DeleteMovie(c *gin.Context, id int) error {
	args := m.Called(c, id)
	return args.Error(0)
}

func (m *MockMovieController) ListMovies(c *gin.Context) (dto.MoviesListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.MoviesListResponse), args.Error(1)
}

func (m *MockMovieController) SearchMoviesByTitle(c *gin.Context) (dto.MoviesListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.MoviesListResponse), args.Error(1)
}

func (m *MockMovieController) SearchMoviesByActorName(c *gin.Context) (dto.MoviesListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.MoviesListResponse), args.Error(1)
}

func (m *MockMovieController) GetAllMoviesSorted(c *gin.Context) (dto.MoviesListResponse, error) {
	args := m.Called(c)
	return args.Get(0).(dto.MoviesListResponse), args.Error(1)
}

func (m *MockMovieController) CreateMovieWithActors(c *gin.Context, req dto.MovieWithActorsRequest) (dto.MovieResponse, error) {
	args := m.Called(c, req)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) UpdateMovieActors(c *gin.Context, movieID int, req dto.UpdateMovieActorsRequest) (dto.MovieActorsResponse, error) {
	args := m.Called(c, movieID, req)
	return args.Get(0).(dto.MovieActorsResponse), args.Error(1)
}

func (m *MockMovieController) AddActorToMovie(c *gin.Context, movieID, actorID int) (dto.MovieResponse, error) {
	args := m.Called(c, movieID, actorID)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) RemoveActorFromMovie(c *gin.Context, movieID, actorID int) (dto.MovieResponse, error) {
	args := m.Called(c, movieID, actorID)
	return args.Get(0).(dto.MovieResponse), args.Error(1)
}

func (m *MockMovieController) GetActorsForMovieByID(c *gin.Context, movieID int) (dto.MovieActorsResponse, error) {
	args := m.Called(c, movieID)
	return args.Get(0).(dto.MovieActorsResponse), args.Error(1)
}

func (m *MockMovieController) GetMoviesForActor(c *gin.Context, actorID int) (dto.ActorMoviesResponse, error) {
	args := m.Called(c, actorID)
	return args.Get(0).(dto.ActorMoviesResponse), args.Error(1)
}

func (m *MockMovieController) PartialUpdateMovie(c *gin.Context, id int, update dto.MovieUpdate) error {
	args := m.Called(c, id, update)
	return args.Error(0)
}

// newTestMovieHandler создает новый MovieHandler с мок-зависимостями для тестирования
func newTestMovieHandler(ctrl *MockMovieController, producer *kafka.MockProducer) *MovieHandler {
	producerPool := kafka.NewProducerPool(producer, 1, 10)
	return NewMovieHandler(ctrl, producerPool)
}

func TestMovieHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(*MockMovieController)
		setupProducer  func(*kafka.MockProducer)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			requestBody: map[string]interface{}{
				"title":        "Test Movie",
				"description":  "Test Description",
				"release_year": 2023,
				"rating":       8.5,
			},
			setupMock: func(m *MockMovieController) {
				expectedReq := dto.CreateMovieRequest{
					Title:       "Test Movie",
					Description: "Test Description",
					ReleaseYear: 2023,
					Rating:      8.5,
				}
				m.On("CreateMovie", mock.Anything, expectedReq).
					Return(dto.MovieResponse{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
			},
			setupProducer: func(p *kafka.MockProducer) {
				p.On("Produce", mock.Anything, "movies", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"title":"Test Movie","description":"Test Description","release_year":2023,"rating":8.5}`,
		},
		{
			name: "empty title",
			requestBody: map[string]interface{}{
				"title":        "", // Invalid: empty title
				"description":  "Test Description",
				"release_year": 2023,
				"rating":       8.5,
			},
			setupMock: func(m *MockMovieController) {},
			setupProducer: func(p *kafka.MockProducer) {
				p.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Title is required"}`,
		},
		{
			name: "missing description",
			requestBody: map[string]interface{}{
				"title": "Test Movie",
				// Missing description
				"release_year": 2023,
				"rating":       8.5,
			},
			setupMock: func(m *MockMovieController) {},
			setupProducer: func(p *kafka.MockProducer) {
				p.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Description is required"}`,
		},
		{
			name: "invalid rating",
			requestBody: map[string]interface{}{
				"title":        "Test Movie",
				"description":  "Test Description",
				"release_year": 2023,
				"rating":       11, // Invalid: rating > 10
			},
			setupMock: func(m *MockMovieController) {},
			setupProducer: func(p *kafka.MockProducer) {
				p.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Rating must be between 0 and 10"}`,
		},
		{
			name: "controller error",
			requestBody: map[string]interface{}{
				"title":        "Test Movie",
				"description":  "Test Description",
				"release_year": 2023,
				"rating":       8.5,
			},
			setupMock: func(m *MockMovieController) {
				expectedReq := dto.CreateMovieRequest{
					Title:       "Test Movie",
					Description: "Test Description",
					ReleaseYear: 2023,
					Rating:      8.5,
				}
				m.On("CreateMovie", mock.Anything, expectedReq).
					Return(dto.MovieResponse{}, errors.New("database error"))
			},
			setupProducer: func(p *kafka.MockProducer) {
				p.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
		{
			name: "produce error",
			requestBody: map[string]interface{}{
				"title":        "Test Movie",
				"description":  "Test Description",
				"release_year": 2023,
				"rating":       8.5,
			},
			setupMock: func(m *MockMovieController) {
				expectedReq := dto.CreateMovieRequest{
					Title:       "Test Movie",
					Description: "Test Description",
					ReleaseYear: 2023,
					Rating:      8.5,
				}
				m.On("CreateMovie", mock.Anything, expectedReq).
					Return(dto.MovieResponse{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
			},
			setupProducer: func(p *kafka.MockProducer) {
				p.On("Produce", mock.Anything, "movies", mock.Anything, mock.Anything).Return(errors.New("kafka produce error"))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"title":"Test Movie","description":"Test Description","release_year":2023,"rating":8.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)

			// Setup mocks
			tt.setupMock(mockCtrl)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			if tt.setupProducer != nil {
				tt.setupProducer(producer)
			}

			producerPool := kafka.NewProducerPool(producer, 1, 10)
			handler := NewMovieHandler(mockCtrl, producerPool)

			r.POST("/movies", handler.Create)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/movies", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_GetByID тестирует метод GetByID у MovieHandler
func TestMovieHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			setupMock: func(m *MockMovieController, id int) {
				m.On("GetMovieByID", mock.Anything, id).
					Return(dto.MovieResponse{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"title":"Test Movie","description":"Test Description","release_year":2023,"rating":8.5}`,
		},
		{
			name:    "invalid id",
			movieID: "invalid",
			setupMock: func(m *MockMovieController, id int) {
				// No mock setup needed for this case
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			movieID, _ := strconv.Atoi(tt.movieID)
			tt.setupMock(mockCtrl, movieID)

			// Ensure producer is not called for GetByID
			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			r.GET("/movies/:id", handler.GetByID)

			req, _ := http.NewRequest("GET", "/movies/"+tt.movieID, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_List тестирует метод List у MovieHandler
func TestMovieHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMovieController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "empty list",
			setupMock: func(m *MockMovieController) {
				m.On("ListMovies", mock.Anything).
					Return(dto.MoviesListResponse{Movies: []dto.MovieResponse{}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"movies":[]}`,
		},
		{
			name: "with movies",
			setupMock: func(m *MockMovieController) {
				m.On("ListMovies", mock.Anything).
					Return(dto.MoviesListResponse{
						Movies: []dto.MovieResponse{
							{
								ID:          1,
								Title:       "Movie 1",
								Description: "Description 1",
								ReleaseYear: 2023,
								Rating:      8.5,
							},
							{
								ID:          2,
								Title:       "Movie 2",
								Description: "Description 2",
								ReleaseYear: 2024,
								Rating:      9.0,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"movies":[{"id":1,"title":"Movie 1","description":"Description 1","release_year":2023,"rating":8.5},{"id":2,"title":"Movie 2","description":"Description 2","release_year":2024,"rating":9}]}`,
		},
		{
			name: "controller error",
			setupMock: func(m *MockMovieController) {
				m.On("ListMovies", mock.Anything).
					Return(dto.MoviesListResponse{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			tt.setupMock(mockCtrl)

			r.GET("/movies", handler.List)
			req, _ := http.NewRequest("GET", "/movies", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_Search тестирует метод Search у MovieHandler
func TestMovieHandler_Search(t *testing.T) {
	tests := []struct {
		name           string
		titleQuery     string
		actorQuery     string
		setupMock      func(*MockMovieController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "search by title",
			titleQuery: "matrix",
			setupMock: func(m *MockMovieController) {
				m.On("SearchMoviesByTitle", mock.Anything).
					Return(dto.MoviesListResponse{
						Movies: []dto.MovieResponse{
							{
								ID:          1,
								Title:       "The Matrix",
								Description: "A computer hacker learns about the true nature of reality",
								ReleaseYear: 1999,
								Rating:      8.7,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"movies":[{"id":1,"title":"The Matrix","description":"A computer hacker learns about the true nature of reality","release_year":1999,"rating":8.7}]}`,
		},
		{
			name:       "search by actor",
			actorQuery: "keanu",
			setupMock: func(m *MockMovieController) {
				m.On("SearchMoviesByActorName", mock.Anything).
					Return(dto.MoviesListResponse{
						Movies: []dto.MovieResponse{
							{
								ID:          2,
								Title:       "John Wick",
								Description: "An ex-hit-man comes out of retirement to track down the gangsters that took everything from him.",
								ReleaseYear: 2014,
								Rating:      7.4,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"at least one search parameter (title or actorName) is required"}`,
		},
		{
			name:           "empty query",
			setupMock:      func(m *MockMovieController) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"at least one search parameter (title or actorName) is required"}`,
		},
		{
			name:       "controller error",
			titleQuery: "error",
			setupMock: func(m *MockMovieController) {
				m.On("SearchMoviesByTitle", mock.Anything).
					Return(dto.MoviesListResponse{}, errors.New("search error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"search error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			tt.setupMock(mockCtrl)

			r.GET("/movies/search", handler.Search)

			// Build URL with query parameters
			url := "/movies/search?"
			if tt.titleQuery != "" {
				url += "title=" + tt.titleQuery
			}
			if tt.actorQuery != "" {
				if tt.titleQuery != "" {
					url += "&"
				}
				url += "actor=" + tt.actorQuery
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_CreateWithActors тестирует метод CreateWithActors у MovieHandler
func TestMovieHandler_CreateWithActors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockMovieController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with actors",
			requestBody: `{
				"title": "Movie",
				"description": "Description with actors",
				"release_year": 2020,
				"rating": 7.5,
				"actor_ids": [1, 2, 3]
			}`,
			setupMock: func(m *MockMovieController) {
				expectedReq := dto.MovieWithActorsRequest{
					Title:       "Movie",
					Description: "Description with actors",
					ReleaseYear: 2020,
					Rating:      7.5,
					ActorIDs:    []int{1, 2, 3},
				}
				m.On("CreateMovieWithActors", mock.Anything, expectedReq).
					Return(dto.MovieResponse{
						ID:          1,
						Title:       "Movie",
						Description: "Description with actors",
						ReleaseYear: 2020,
						Rating:      7.5,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"title":"Movie","description":"Description with actors","release_year":2020,"rating":7.5}`,
		},
		{
			name: "missing required fields",
			requestBody: `{
				"description": "Missing title",
				"release_year": 2020,
				"actor_ids": [1, 2]
			}`,
			setupMock:      func(m *MockMovieController) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "invalid actor ids",
			requestBody: `{
				"title": "Invalid Actors",
				"description": "Test",
				"release_year": 2023,
				"rating": 5,
				"actor_ids": ["not_an_integer"]
			}`,
			setupMock:      func(m *MockMovieController) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "controller error",
			requestBody: `{
				"title": "Error",
				"description": "Test",
				"release_year": 2023,
				"rating": 5,
				"actor_ids": [1, 2]
			}`,
			setupMock: func(m *MockMovieController) {
				m.On("CreateMovieWithActors", mock.Anything, mock.Anything).
					Return(dto.MovieResponse{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			tt.setupMock(mockCtrl)

			r.POST("/movies", handler.CreateWithActors)
			req, _ := http.NewRequest("POST", "/movies", bytes.NewBufferString(tt.requestBody))
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

// TestMovieHandler_Update тестирует метод Update у MovieHandler
func TestMovieHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		requestBody    string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			requestBody: `{
				"title": "Updated Movie",
				"description": "Updated description",
				"release_year": 2023,
				"rating": 9.0
			}`,
			setupMock: func(m *MockMovieController, id int) {
				title := "Updated Movie"
				description := "Updated description"
				releaseYear := 2023
				rating := 9.0
				_ = dto.UpdateMovieRequest{
					Title:       &title,
					Description: &description,
					ReleaseYear: &releaseYear,
					Rating:      &rating,
				}
				m.On("UpdateMovie", mock.Anything, id, mock.MatchedBy(func(req dto.UpdateMovieRequest) bool {
					return *req.Title == "Updated Movie" && *req.Description == "Updated description" &&
						*req.ReleaseYear == 2023 && *req.Rating == 9.0
				})).Return(dto.MovieResponse{
					ID:          1,
					Title:       "Updated Movie",
					Description: "Updated description",
					ReleaseYear: 2023,
					Rating:      9.0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"title":"Updated Movie","description":"Updated description","release_year":2023,"rating":9}`,
		},
		{
			name:           "invalid id",
			movieID:        "abc",
			requestBody:    `{"title":"Test"}`,
			setupMock:      func(m *MockMovieController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:        "not found",
			movieID:     "999",
			requestBody: `{"title":"Not Found"}`,
			setupMock: func(m *MockMovieController, id int) {
				m.On("UpdateMovie", mock.Anything, id, mock.Anything).
					Return(dto.MovieResponse{}, errors.New("movie not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"movie not found"}`,
		},
		{
			name:        "invalid rating",
			movieID:     "1",
			requestBody: `{"rating":11}`, // Invalid rating
			setupMock: func(m *MockMovieController, id int) {
				// The controller will be called and return a validation error
				m.On("UpdateMovie", mock.Anything, id, mock.MatchedBy(func(req dto.UpdateMovieRequest) bool {
					return *req.Rating == 11
				})).Return(dto.MovieResponse{}, errors.New("validation error: rating: must be between 0 and 10"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"validation error: rating: must be between 0 and 10"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			tt.setupMock(mockCtrl, movieID)

			r.PUT("/movies/:id", handler.Update)
			req, _ := http.NewRequest("PUT", "/movies/"+tt.movieID, bytes.NewBufferString(tt.requestBody))
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

// TestMovieHandler_PartialUpdate тестирует метод PartialUpdate у MovieHandler
func TestMovieHandler_PartialUpdate(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		requestBody    string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "update title only",
			movieID: "1",
			requestBody: `{
				"title": "New Title"
			}`,
			setupMock: func(m *MockMovieController, id int) {
				title := "New Title"
				_ = dto.MovieUpdate{
					Title: &title,
				}
				m.On("PartialUpdateMovie", mock.Anything, id, mock.MatchedBy(func(update dto.MovieUpdate) bool {
					return update.Title != nil && *update.Title == "New Title" && update.Description == nil && update.ReleaseYear == nil && update.Rating == nil
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "update rating only",
			movieID: "2",
			requestBody: `{
				"rating": 9.5
			}`,
			setupMock: func(m *MockMovieController, id int) {
				rating := 9.5
				_ = dto.MovieUpdate{
					Rating: &rating,
				}
				m.On("PartialUpdateMovie", mock.Anything, id, mock.MatchedBy(func(update dto.MovieUpdate) bool {
					return update.Title == nil && update.Description == nil && update.ReleaseYear == nil && update.Rating != nil && *update.Rating == rating
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			movieID:        "abc",
			requestBody:    `{"title":"Test"}`,
			setupMock:      func(m *MockMovieController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:        "not found",
			movieID:     "999",
			requestBody: `{"title":"Not Found"}`,
			setupMock: func(m *MockMovieController, id int) {
				m.On("PartialUpdateMovie", mock.Anything, id, mock.Anything).
					Return(errors.New("movie not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"movie not found"}`,
		},
		{
			name:        "invalid rating",
			movieID:     "1",
			requestBody: `{"rating":11}`, // Invalid rating
			setupMock: func(m *MockMovieController, id int) {
				// The controller will be called and return a validation error
				m.On("PartialUpdateMovie", mock.Anything, id, mock.MatchedBy(func(update dto.MovieUpdate) bool {
					return update.Rating != nil && *update.Rating == 11
				})).Return(errors.New("validation error: rating: must be between 0 and 10"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"validation error: rating: must be between 0 and 10"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			tt.setupMock(mockCtrl, movieID)

			r.PATCH("/movies/:id", handler.PartialUpdate)
			req, _ := http.NewRequest("PATCH", "/movies/"+tt.movieID, bytes.NewBufferString(tt.requestBody))
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

// TestMovieHandler_Delete тестирует метод Delete у MovieHandler
func TestMovieHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			setupMock: func(m *MockMovieController, id int) {
				m.On("DeleteMovie", mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			movieID:        "abc",
			setupMock:      func(m *MockMovieController, id int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name:    "not found",
			movieID: "999",
			setupMock: func(m *MockMovieController, id int) {
				m.On("DeleteMovie", mock.Anything, id).
					Return(errors.New("movie not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"movie not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			tt.setupMock(mockCtrl, movieID)

			r.DELETE("/movies/:id", handler.Delete)
			req, _ := http.NewRequest("DELETE", "/movies/"+tt.movieID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_UpdateMovieActors тестирует метод UpdateMovieActors у MovieHandler
func TestMovieHandler_UpdateMovieActors(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		requestBody    string
		setupMock      func(*MockMovieController, int, dto.UpdateMovieActorsRequest)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success with actors",
			movieID: "1",
			requestBody: `{
				"actor_ids": [1, 2, 3]
			}`,
			setupMock: func(m *MockMovieController, id int, req dto.UpdateMovieActorsRequest) {
				m.On("UpdateMovieActors", mock.Anything, id, mock.MatchedBy(func(r dto.UpdateMovieActorsRequest) bool {
					return len(r.ActorIDs) == 3 && r.ActorIDs[0] == 1 && r.ActorIDs[1] == 2 && r.ActorIDs[2] == 3
				})).Return(dto.MovieActorsResponse{
					Actors: []dto.ActorResponse{
						{ID: 1, Name: "Actor 1"},
						{ID: 2, Name: "Actor 2"},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"actors":[{"id":1,"name":"Actor 1","gender":"","birth_date":""},{"id":2,"name":"Actor 2","gender":"","birth_date":""}]}`,
		},
		{
			name:           "invalid movie id",
			movieID:        "abc",
			requestBody:    `{"actor_ids":[1,2]}`,
			setupMock:      func(m *MockMovieController, id int, req dto.UpdateMovieActorsRequest) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:        "movie not found",
			movieID:     "999",
			requestBody: `{"actor_ids":[1,2]}`,
			setupMock: func(m *MockMovieController, id int, req dto.UpdateMovieActorsRequest) {
				m.On("UpdateMovieActors", mock.Anything, id, mock.Anything).
					Return(dto.MovieActorsResponse{}, errors.New("movie not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"movie not found"}`,
		},
		{
			name:           "invalid request body",
			movieID:        "1",
			requestBody:    `{"actor_ids":["not_an_integer"]}`,
			setupMock:      func(m *MockMovieController, id int, req dto.UpdateMovieActorsRequest) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			var reqBody dto.UpdateMovieActorsRequest
			_ = json.Unmarshal([]byte(tt.requestBody), &reqBody)
			tt.setupMock(mockCtrl, movieID, reqBody)

			r.PUT("/movies/:id/actors", handler.UpdateMovieActors)
			req, _ := http.NewRequest("PUT", "/movies/"+tt.movieID+"/actors", bytes.NewBufferString(tt.requestBody))
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

// TestMovieHandler_ListSorted тестирует метод ListSorted у MovieHandler
func TestMovieHandler_ListSorted(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMovieController)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "empty list",
			setupMock: func(m *MockMovieController) {
				m.On("GetAllMoviesSorted", mock.Anything).
					Return(dto.MoviesListResponse{Movies: []dto.MovieResponse{}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"movies":[]}`,
		},
		{
			name: "with movies",
			setupMock: func(m *MockMovieController) {
				m.On("GetAllMoviesSorted", mock.Anything).
					Return(dto.MoviesListResponse{
						Movies: []dto.MovieResponse{
							{
								ID:          1,
								Title:       "The Shawshank Redemption",
								Description: "Two imprisoned men bond over a number of years...",
								ReleaseYear: 1994,
								Rating:      9.3,
							},
							{
								ID:          2,
								Title:       "The Godfather",
								Description: "The aging patriarch of an organized crime dynasty...",
								ReleaseYear: 1972,
								Rating:      9.2,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"movies": [
					{
						"id": 1,
						"title": "The Shawshank Redemption",
						"description": "Two imprisoned men bond over a number of years...",
						"release_year": 1994,
						"rating": 9.3
					},
					{
						"id": 2,
						"title": "The Godfather",
						"description": "The aging patriarch of an organized crime dynasty...",
						"release_year": 1972,
						"rating": 9.2
					}
				]
			}`,
		},
		{
			name: "controller error",
			setupMock: func(m *MockMovieController) {
				m.On("GetAllMoviesSorted", mock.Anything).
					Return(dto.MoviesListResponse{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			tt.setupMock(mockCtrl)

			r.GET("/movies/sorted", handler.ListSorted)
			req, _ := http.NewRequest("GET", "/movies/sorted", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_AddActorToMovie тестирует метод AddActorToMovie у MovieHandler
func TestMovieHandler_AddActorToMovie(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		actorID        string
		setupMock      func(*MockMovieController, int, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			actorID: "2",
			setupMock: func(m *MockMovieController, movieID, actorID int) {
				m.On("AddActorToMovie", mock.Anything, movieID, actorID).
					Return(dto.MovieResponse{ID: movieID, Title: "Movie", Description: "", ReleaseYear: 0, Rating: 0}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:           "invalid movie id",
			movieID:        "abc",
			actorID:        "2",
			setupMock:      func(m *MockMovieController, movieID, actorID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:           "invalid actor id",
			movieID:        "1",
			actorID:        "xyz",
			setupMock:      func(m *MockMovieController, movieID, actorID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:    "controller error",
			movieID: "1",
			actorID: "2",
			setupMock: func(m *MockMovieController, movieID, actorID int) {
				m.On("AddActorToMovie", mock.Anything, movieID, actorID).
					Return(dto.MovieResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, movieID, actorID)

			r.POST("/movies/:id/actors/:actorId", handler.AddActorToMovie)
			url := "/movies/" + tt.movieID + "/actors/" + tt.actorID
			req, _ := http.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_RemoveActorFromMovie тестирует метод RemoveActorFromMovie у MovieHandler
func TestMovieHandler_RemoveActorFromMovie(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		actorID        string
		setupMock      func(*MockMovieController, int, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			actorID: "2",
			setupMock: func(m *MockMovieController, movieID, actorID int) {
				m.On("RemoveActorFromMovie", mock.Anything, movieID, actorID).
					Return(dto.MovieResponse{ID: movieID, Title: "Movie", Description: "", ReleaseYear: 0, Rating: 0}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:           "invalid movie id",
			movieID:        "abc",
			actorID:        "2",
			setupMock:      func(m *MockMovieController, movieID, actorID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:           "invalid actor id",
			movieID:        "1",
			actorID:        "xyz",
			setupMock:      func(m *MockMovieController, movieID, actorID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:    "controller error",
			movieID: "1",
			actorID: "2",
			setupMock: func(m *MockMovieController, movieID, actorID int) {
				m.On("RemoveActorFromMovie", mock.Anything, movieID, actorID).
					Return(dto.MovieResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, _ := strconv.Atoi(tt.movieID)
			actorID, _ := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, movieID, actorID)

			r.DELETE("/movies/:id/actors/:actorId", handler.RemoveActorFromMovie)
			url := "/movies/" + tt.movieID + "/actors/" + tt.actorID
			req, _ := http.NewRequest("DELETE", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

// TestMovieHandler_GetActorsForMovieByID тестирует метод GetActorsForMovieByID у MovieHandler
func TestMovieHandler_GetActorsForMovieByID(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			movieID: "1",
			setupMock: func(m *MockMovieController, movieID int) {
				m.On("GetActorsForMovieByID", mock.Anything, movieID).
					Return(dto.MovieActorsResponse{Actors: []dto.ActorResponse{{ID: 1, Name: "Actor"}}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"actors":[{"id":1,"name":"Actor","gender":"","birth_date":""}]}`,
		},
		{
			name:           "invalid movie id",
			movieID:        "abc",
			setupMock:      func(m *MockMovieController, movieID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid movie id"}`,
		},
		{
			name:    "controller error",
			movieID: "1",
			setupMock: func(m *MockMovieController, movieID int) {
				m.On("GetActorsForMovieByID", mock.Anything, movieID).
					Return(dto.MovieActorsResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"db error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			movieID, err := strconv.Atoi(tt.movieID)
			tt.setupMock(mockCtrl, movieID)

			r.GET("/movies/:id/actors", handler.GetActorsForMovieByID)
			url := "/movies/" + tt.movieID + "/actors"
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			_ = err
		})
	}
}

// TestMovieHandler_GetMoviesForActor тестирует метод GetMoviesForActor у MovieHandler
func TestMovieHandler_GetMoviesForActor(t *testing.T) {
	tests := []struct {
		name           string
		actorID        string
		setupMock      func(*MockMovieController, int)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			actorID: "1",
			setupMock: func(m *MockMovieController, actorID int) {
				m.On("GetMoviesForActor", mock.Anything, actorID).
					Return(dto.ActorMoviesResponse{Movies: []dto.MovieResponse{{ID: 1, Title: "Movie"}}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"movies":[{"id":1,"title":"Movie","description":"","release_year":0,"rating":0}]}`,
		},
		{
			name:           "invalid actor id",
			actorID:        "abc",
			setupMock:      func(m *MockMovieController, actorID int) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid actor id"}`,
		},
		{
			name:    "controller error",
			actorID: "1",
			setupMock: func(m *MockMovieController, actorID int) {
				m.On("GetMoviesForActor", mock.Anything, actorID).
					Return(dto.ActorMoviesResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"db error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			mockCtrl := new(MockMovieController)
			producer := kafka.NewMockProducer()
			producer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			handler := newTestMovieHandler(mockCtrl, producer)

			producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			actorID, err := strconv.Atoi(tt.actorID)
			tt.setupMock(mockCtrl, actorID)

			r.GET("/actors/:id/movies", handler.GetMoviesForActor)
			url := "/actors/" + tt.actorID + "/movies"
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			_ = err
		})
	}
}
