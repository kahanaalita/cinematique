package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cinematique/internal/controller/dto"
	"cinematique/internal/domain"
)

// Вспомогательная функция для создания указателей на string, int, float64
func ptr[T any](v T) *T { return &v }

// MockMovieService - мок сервиса фильмов
type MockMovieService struct {
	mock.Mock
}

func (m *MockMovieService) Create(movie domain.Movie, actorIDs []int) (int, error) {
	args := m.Called(movie, actorIDs)
	return args.Int(0), args.Error(1)
}

func (m *MockMovieService) GetByID(id int) (domain.Movie, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Movie), args.Error(1)
}

func (m *MockMovieService) Update(movie domain.Movie, actorIDs []int) error {
	args := m.Called(movie, actorIDs)
	return args.Error(0)
}

func (m *MockMovieService) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMovieService) GetAll() ([]domain.Movie, error) {
	args := m.Called()
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func (m *MockMovieService) AddActor(movieID, actorID int) error {
	args := m.Called(movieID, actorID)
	return args.Error(0)
}

func (m *MockMovieService) RemoveActor(movieID, actorID int) error {
	args := m.Called(movieID, actorID)
	return args.Error(0)
}

func (m *MockMovieService) GetActors(movieID int) ([]domain.Actor, error) {
	args := m.Called(movieID)
	return args.Get(0).([]domain.Actor), args.Error(1)
}

func (m *MockMovieService) GetActorsForMovieByID(movieID int) ([]domain.Actor, error) {
	args := m.Called(movieID)
	return args.Get(0).([]domain.Actor), args.Error(1)
}

func (m *MockMovieService) GetMoviesForActor(actorID int) ([]domain.Movie, error) {
	args := m.Called(actorID)
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func (m *MockMovieService) SearchMoviesByTitle(titleFragment string) ([]domain.Movie, error) {
	args := m.Called(titleFragment)
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func (m *MockMovieService) SearchMoviesByActorName(actorNameFragment string) ([]domain.Movie, error) {
	args := m.Called(actorNameFragment)
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func (m *MockMovieService) GetAllMoviesSorted(sortField, sortOrder string) ([]domain.Movie, error) {
	args := m.Called(sortField, sortOrder)
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func (m *MockMovieService) CreateMovieWithActors(movie domain.Movie, actorIDs []int) (int, error) {
	args := m.Called(movie, actorIDs)
	return args.Int(0), args.Error(1)
}

func (m *MockMovieService) UpdateMovieActors(movieID int, actorIDs []int) error {
	args := m.Called(movieID, actorIDs)
	return args.Error(0)
}

func (m *MockMovieService) PartialUpdateMovie(id int, update domain.MovieUpdate) error {
	args := m.Called(id, update)
	return args.Error(0)
}

func TestMovieController_CreateMovie(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.CreateMovieRequest
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name: "success",
			req: dto.CreateMovieRequest{
				Title:       "Test Movie",
				Description: "Test Description",
				ReleaseYear: 2023,
				Rating:      8.5,
				ActorIDs:    []int{1, 2},
			},
			setupMock: func(mms *MockMovieService) {
				mms.On("Create", mock.AnythingOfType("domain.Movie"), []int{1, 2}).
					Return(1, nil)
				mms.On("GetByID", 1).
					Return(domain.Movie{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
			},
			expectedError: false,
		},
		{
			name: "invalid rating",
			req: dto.CreateMovieRequest{
				Title:       "Test Movie",
				Description: "Test Description",
				ReleaseYear: 2023,
				Rating:      11.0, // Invalid rating
			},
			setupMock:     func(mms *MockMovieService) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			_, err := controller.CreateMovie(&gin.Context{}, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_GetMovieByID(t *testing.T) {
	tests := []struct {
		name          string
		movieID       int
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name:    "success",
			movieID: 1,
			setupMock: func(mms *MockMovieService) {
				mms.On("GetByID", 1).
					Return(domain.Movie{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
			},
			expectedError: false,
		},
		{
			name:    "not found",
			movieID: 999,
			setupMock: func(mms *MockMovieService) {
				mms.On("GetByID", 999).
					Return(domain.Movie{}, errors.New("movie not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			_, err := controller.GetMovieByID(&gin.Context{}, tt.movieID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_UpdateMovie(t *testing.T) {
	movieID := 1
	tests := []struct {
		name          string
		req           dto.UpdateMovieRequest
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name: "success",
			req: dto.UpdateMovieRequest{
				Title:       ptr("Updated Movie"),
				Description: ptr("Updated Description"),
				ReleaseYear: ptr(2024),
				Rating:      ptr(9.0),
				ActorIDs:    &[]int{1, 2, 3},
			},
			setupMock: func(mms *MockMovieService) {
				mms.On("GetByID", movieID).
					Return(domain.Movie{
						ID:          movieID,
						Title:       "Original Movie",
						Description: "Original Description",
						ReleaseYear: 2023,
						Rating:      8.5,
					}, nil)
				mms.On("Update", mock.MatchedBy(func(movie domain.Movie) bool {
					return movie.Title == "Updated Movie" &&
						movie.Description == "Updated Description" &&
						movie.ReleaseYear == 2024 &&
						movie.Rating == 9.0
				}), []int{1, 2, 3}).Return(nil)
				mms.On("GetByID", movieID).
					Return(domain.Movie{
						ID:          movieID,
						Title:       "Updated Movie",
						Description: "Updated Description",
						ReleaseYear: 2024,
						Rating:      9.0,
					}, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			_, err := controller.UpdateMovie(&gin.Context{}, movieID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_DeleteMovie(t *testing.T) {
	tests := []struct {
		name          string
		movieID       int
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name:    "success",
			movieID: 1,
			setupMock: func(mms *MockMovieService) {
				mms.On("Delete", 1).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "not found",
			movieID: 999,
			setupMock: func(mms *MockMovieService) {
				mms.On("Delete", 999).Return(errors.New("movie not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			err := controller.DeleteMovie(&gin.Context{}, tt.movieID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_ListMovies(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMovieService)
		expectedResult dto.MoviesListResponse
		expectedError  bool
	}{
		{
			name: "success",
			setupMock: func(mms *MockMovieService) {
				mms.On("GetAll").Return([]domain.Movie{
					{
						ID:          1,
						Title:       "Movie 1",
						Description: "Description 1",
						ReleaseYear: 2020,
						Rating:      8.5,
					},
				}, nil)
			},
			expectedResult: dto.MoviesListResponse{
				Movies: []dto.MovieResponse{
					{
						ID:          1,
						Title:       "Movie 1",
						Description: "Description 1",
						ReleaseYear: 2020,
						Rating:      8.5,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "empty list",
			setupMock: func(mms *MockMovieService) {
				mms.On("GetAll").Return([]domain.Movie{}, nil)
			},
			expectedResult: dto.MoviesListResponse{
				Movies: []dto.MovieResponse{},
			},
			expectedError: false,
		},
		{
			name: "service error",
			setupMock: func(mms *MockMovieService) {
				mms.On("GetAll").Return([]domain.Movie{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			result, err := controller.ListMovies(&gin.Context{})

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_SearchMoviesByTitle(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(*MockMovieService)
		expectedResult dto.MoviesListResponse
		expectedError  bool
	}{
		{
			name:  "success",
			query: "test",
			setupMock: func(mms *MockMovieService) {
				mms.On("SearchMoviesByTitle", "test").Return([]domain.Movie{
					{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      7.5,
					},
				}, nil)
			},
			expectedResult: dto.MoviesListResponse{
				Movies: []dto.MovieResponse{
					{
						ID:          1,
						Title:       "Test Movie",
						Description: "Test Description",
						ReleaseYear: 2023,
						Rating:      7.5,
					},
				},
			},
			expectedError: false,
		},
		{
			name:      "empty_query",
			query:     "",
			setupMock: nil,
			expectedResult: dto.MoviesListResponse{
				Movies: []dto.MovieResponse{},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			controller := NewMovieController(mockService)

			// Устанавливаем query параметр 'title', а не 'query'
			ctx := &gin.Context{}
			ctx.Request = &http.Request{
				URL: &url.URL{RawQuery: "title=" + tt.query},
			}

			result, err := controller.SearchMoviesByTitle(ctx)

			if tt.name == "empty_query" {
				assert.Error(t, err)
				assert.EqualError(t, err, "title parameter is required")
			} else if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			if tt.setupMock != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestMovieController_GetAllMoviesSorted(t *testing.T) {
	tests := []struct {
		name           string
		sortField      string
		sortOrder      string
		setupMock      func(*MockMovieService)
		expectedResult dto.MoviesListResponse
		expectedError  bool
	}{
		{
			name:      "default sort by rating desc",
			sortField: "",
			sortOrder: "",
			setupMock: func(mms *MockMovieService) {
				mms.On("GetAllMoviesSorted", "rating", "DESC").Return([]domain.Movie{
					{
						ID:          1,
						Title:       "A Movie",
						Description: "Description",
						ReleaseYear: 2020,
						Rating:      8.0,
					},
				}, nil)
			},
			expectedResult: dto.MoviesListResponse{
				Movies: []dto.MovieResponse{
					{
						ID:          1,
						Title:       "A Movie",
						Description: "Description",
						ReleaseYear: 2020,
						Rating:      8.0,
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			// Устанавливаем query параметры
			ctx := &gin.Context{}
			ctx.Request = &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("sort_by=%s&order=%s", tt.sortField, tt.sortOrder),
				},
			}

			result, err := controller.GetAllMoviesSorted(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_CreateMovieWithActors(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.MovieWithActorsRequest
		setupMock     func(*MockMovieService)
		expectedID    int
		expectedError bool
	}{
		{
			name: "success",
			req: dto.MovieWithActorsRequest{
				Title:       "New Movie",
				Description: "Description",
				ReleaseYear: 2023,
				Rating:      8.5,
				ActorIDs:    []int{1, 2},
			},
			setupMock: func(mms *MockMovieService) {
				mms.On("CreateMovieWithActors", domain.Movie{
					Title:       "New Movie",
					Description: "Description",
					ReleaseYear: 2023,
					Rating:      8.5,
				}, []int{1, 2}).Return(1, nil)

				// Add mock for GetByID call
				mms.On("GetByID", 1).Return(domain.Movie{
					ID:          1,
					Title:       "New Movie",
					Description: "Description",
					ReleaseYear: 2023,
					Rating:      8.5,
				}, nil)
			},
			expectedID:    1,
			expectedError: false,
		},
		{
			name: "validation error",
			req: dto.MovieWithActorsRequest{
				Title:       "", // Пустое название
				Description: "Description",
				ReleaseYear: 2023,
				Rating:      11.0, // Некорректный рейтинг
				ActorIDs:    []int{1, 2},
			},
			setupMock:     func(mms *MockMovieService) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			result, err := controller.CreateMovieWithActors(&gin.Context{}, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, result.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_AddActorToMovie(t *testing.T) {
	tests := []struct {
		name          string
		movieID       int
		actorID       int
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name:    "success",
			movieID: 1,
			actorID: 2,
			setupMock: func(mms *MockMovieService) {
				mms.On("AddActor", 1, 2).Return(nil)
				mms.On("GetByID", 1).Return(domain.Movie{
					ID:          1,
					Title:       "Test Movie",
					Description: "Test Description",
					ReleaseYear: 2023,
					Rating:      7.5,
				}, nil)
			},
			expectedError: false,
		},
		{
			name:    "movie not found",
			movieID: 999,
			actorID: 1,
			setupMock: func(mms *MockMovieService) {
				mms.On("AddActor", 999, 1).Return(errors.New("movie not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			_, err := controller.AddActorToMovie(&gin.Context{}, tt.movieID, tt.actorID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_RemoveActorFromMovie(t *testing.T) {
	tests := []struct {
		name          string
		movieID       int
		actorID       int
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name:    "success",
			movieID: 1,
			actorID: 2,
			setupMock: func(mms *MockMovieService) {
				mms.On("RemoveActor", 1, 2).Return(nil)
				mms.On("GetByID", 1).Return(domain.Movie{
					ID:          1,
					Title:       "Test Movie",
					Description: "Test Description",
					ReleaseYear: 2023,
					Rating:      7.5,
				}, nil)
			},
			expectedError: false,
		},
		{
			name:    "movie not found",
			movieID: 999,
			actorID: 1,
			setupMock: func(mms *MockMovieService) {
				mms.On("RemoveActor", 999, 1).Return(errors.New("movie not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			_, err := controller.RemoveActorFromMovie(&gin.Context{}, tt.movieID, tt.actorID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_GetActorsForMovieByID(t *testing.T) {
	tests := []struct {
		name           string
		movieID        int
		setupMock      func(*MockMovieService)
		expectedResult dto.MovieActorsResponse
		expectedError  bool
	}{
		{
			name:    "success",
			movieID: 1,
			setupMock: func(mms *MockMovieService) {
				// Настраиваем ожидание вызова GetByID для проверки существования фильма
				mms.On("GetByID", 1).Return(domain.Movie{ID: 1}, nil)
				// Настраиваем ожидание вызова GetActorsForMovieByID
				mms.On("GetActorsForMovieByID", 1).Return([]domain.Actor{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
			},
			expectedResult: dto.MovieActorsResponse{
				Actors: []dto.ActorResponse{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: "1990-01-01",
					},
				},
			},
			expectedError: false,
		},
		{
			name:    "movie not found",
			movieID: 999,
			setupMock: func(mms *MockMovieService) {
				// Настраиваем ожидание вызова GetByID, возвращаем ошибку "movie not found"
				mms.On("GetByID", 999).Return(domain.Movie{}, domain.ErrMovieNotFound)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			result, err := controller.GetActorsForMovieByID(&gin.Context{}, tt.movieID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_GetMoviesForActor(t *testing.T) {
	tests := []struct {
		name           string
		actorID        int
		setupMock      func(*MockMovieService)
		expectedResult dto.ActorMoviesResponse
		expectedError  bool
	}{
		{
			name:    "success",
			actorID: 1,
			setupMock: func(mms *MockMovieService) {
				mms.On("GetMoviesForActor", 1).Return([]domain.Movie{
					{
						ID:          1,
						Title:       "Movie 1",
						Description: "Description 1",
						ReleaseYear: 2020,
						Rating:      8.5,
					},
				}, nil)
			},
			expectedResult: dto.ActorMoviesResponse{
				Movies: []dto.MovieResponse{
					{
						ID:          1,
						Title:       "Movie 1",
						Description: "Description 1",
						ReleaseYear: 2020,
						Rating:      8.5,
					},
				},
			},
			expectedError: false,
		},
		{
			name:    "actor not found",
			actorID: 999,
			setupMock: func(mms *MockMovieService) {
				mms.On("GetMoviesForActor", 999).Return([]domain.Movie{}, errors.New("actor not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			result, err := controller.GetMoviesForActor(&gin.Context{}, tt.actorID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMovieController_PartialUpdateMovie(t *testing.T) {
	tests := []struct {
		name          string
		movieID       int
		update        dto.MovieUpdate
		setupMock     func(*MockMovieService)
		expectedError bool
	}{
		{
			name:    "update title",
			movieID: 1,
			update: dto.MovieUpdate{
				Title: ptr("Updated Title"),
			},
			setupMock: func(mms *MockMovieService) {
				mms.On("GetByID", 1).Return(domain.Movie{
					ID:          1,
					Title:       "Old Title",
					Description: "Description",
					ReleaseYear: 2020,
					Rating:      8.0,
				}, nil)
				mms.On("Update", domain.Movie{
					ID:          1,
					Title:       "Updated Title",
					Description: "Description",
					ReleaseYear: 2020,
					Rating:      8.0,
				}, []int{}).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "movie not found",
			movieID: 999,
			update: dto.MovieUpdate{
				Title: ptr("New Title"),
			},
			setupMock: func(mms *MockMovieService) {
				mms.On("GetByID", 999).Return(domain.Movie{}, errors.New("movie not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMovieService{}
			tt.setupMock(mockService)

			controller := NewMovieController(mockService)

			err := controller.PartialUpdateMovie(&gin.Context{}, tt.movieID, tt.update)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}
