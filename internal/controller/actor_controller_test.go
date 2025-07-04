package controller

import (
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cinematique/internal/domain"
	"cinematique/internal/controller/dto"
)

// MockActorService - мок сервиса актеров
type MockActorService struct {
	mock.Mock
}

func (m *MockActorService) Create(actor domain.Actor) (int, error) {
	args := m.Called(actor)
	return args.Int(0), args.Error(1)
}

func (m *MockActorService) GetByID(id int) (domain.Actor, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Actor), args.Error(1)
}

func (m *MockActorService) Update(actor domain.Actor) error {
	args := m.Called(actor)
	return args.Error(0)
}

func (m *MockActorService) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockActorService) GetAll() ([]domain.Actor, error) {
	args := m.Called()
	return args.Get(0).([]domain.Actor), args.Error(1)
}

func (m *MockActorService) GetAllActorsWithMovies() ([]domain.Actor, error) {
	args := m.Called()
	return args.Get(0).([]domain.Actor), args.Error(1)
}

func (m *MockActorService) GetMovies(actorID int) ([]domain.Movie, error) {
	args := m.Called(actorID)
	return args.Get(0).([]domain.Movie), args.Error(1)
}

func TestActorController_CreateActor(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.CreateActorRequest
		setupMock     func(*MockActorService)
		expectedError bool
	}{
		{
			name: "success",
			req: dto.CreateActorRequest{
				Name:      "Test Actor",
				Gender:    "male",
				BirthDate: "1990-01-01",
			},
			setupMock: func(mas *MockActorService) {
				mas.On("Create", mock.AnythingOfType("domain.Actor")).
					Return(1, nil)
			},
			expectedError: false,
		},
		{
			name: "invalid gender",
			req: dto.CreateActorRequest{
				Name:      "Test Actor",
				Gender:    "invalid",
				BirthDate: "1990-01-01",
			},
			setupMock:     func(mas *MockActorService) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			_, err := controller.CreateActor(&gin.Context{}, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestActorController_GetActorByID(t *testing.T) {
	tests := []struct {
		name          string
		actorID       int
		setupMock     func(*MockActorService)
		expectedError bool
	}{
		{
			name:    "success",
			actorID: 1,
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 1).
					Return(domain.Actor{
						ID:        1,
						Name:      "Test Actor",
						Gender:    "male",
						BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
			},
			expectedError: false,
		},
		{
			name:    "not found",
			actorID: 999,
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 999).
					Return(domain.Actor{}, errors.New("actor not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			_, err := controller.GetActorByID(&gin.Context{}, tt.actorID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestActorController_UpdateActor(t *testing.T) {
	actorID := 1
	birthDate := "1990-01-01"
	birthTime, _ := time.Parse("2006-01-02", birthDate)
	
	tests := []struct {
		name          string
		req           dto.UpdateActorRequest
		setupMock     func(*MockActorService)
		expectedError bool
	}{
		{
			name: "success",
			req: dto.UpdateActorRequest{
				Name:      stringPtr("Updated Actor"),
				Gender:    stringPtr("female"),
				BirthDate: &birthDate,
			},
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", actorID).
					Return(domain.Actor{
						ID:        actorID,
						Name:      "Original Actor",
						Gender:    "male",
						BirthDate: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				mas.On("Update", mock.MatchedBy(func(actor domain.Actor) bool {
					return actor.Name == "Updated Actor" && 
					       actor.Gender == "female" && 
					       actor.BirthDate.Equal(birthTime)
				})).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			_, err := controller.UpdateActor(&gin.Context{}, actorID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestActorController_DeleteActor(t *testing.T) {
	tests := []struct {
		name          string
		actorID       int
		setupMock     func(*MockActorService)
		expectedError bool
	}{
		{
			name:    "success",
			actorID: 1,
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 1).Return(domain.Actor{ID: 1}, nil)
				mas.On("GetMovies", 1).Return([]domain.Movie{}, nil)
				mas.On("Delete", 1).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "actor has movies",
			actorID: 1,
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 1).Return(domain.Actor{ID: 1}, nil)
				mas.On("GetMovies", 1).Return([]domain.Movie{{ID: 1, Title: "Test Movie"}}, nil)
				// Delete не должен вызываться, если у актера есть фильмы
				mas.AssertNotCalled(t, "Delete", 1)
			},
			expectedError: true,
		},
		{
			name:    "not found",
			actorID: 999,
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 999).Return(domain.Actor{}, domain.ErrActorNotFound)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			err := controller.DeleteActor(&gin.Context{}, tt.actorID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}

// Вспомогательная функция для создания указателя на int
func intPtr(i int) *int {
	return &i
}

// Вспомогательная функция для создания указателя на float64
func float64Ptr(f float64) *float64 {
	return &f
}

func TestActorController_PartialUpdateActor(t *testing.T) {
	tests := []struct {
		name          string
		actorID       int
		update        dto.ActorUpdate
		setupMock     func(*MockActorService)
		expectedError bool
	}{
		{
			name:    "success update name",
			actorID: 1,
			update: dto.ActorUpdate{
				Name: stringPtr("Updated Name"),
			},
			setupMock: func(mas *MockActorService) {
				// Ожидаем вызов GetByID и возвращаем существующего актера
				mas.On("GetByID", 1).Return(domain.Actor{
					ID:        1,
					Name:      "Old Name",
					Gender:    "male",
					BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				// Ожидаем вызов Update с обновленным именем
				mas.On("Update", domain.Actor{
					ID:        1,
					Name:      "Updated Name",
					Gender:    "male",
					BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				}).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "actor not found",
			actorID: 999,
			update: dto.ActorUpdate{
				Name: stringPtr("New Name"),
			},
			setupMock: func(mas *MockActorService) {
				mas.On("GetByID", 999).Return(domain.Actor{}, errors.New("actor not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			_, err := controller.PartialUpdateActor(&gin.Context{}, tt.actorID, tt.update)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestActorController_ListActors(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockActorService)
		expectedResult dto.ActorsListResponse
		expectedError  bool
	}{
		{
			name: "success",
			setupMock: func(mas *MockActorService) {
				mas.On("GetAll").Return([]domain.Actor{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        2,
						Name:      "Actor 2",
						Gender:    "female",
						BirthDate: time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
			},
			expectedResult: dto.ActorsListResponse{
				Actors: []dto.ActorResponse{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: "1990-01-01",
					},
					{
						ID:        2,
						Name:      "Actor 2",
						Gender:    "female",
						BirthDate: "1995-05-05",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "empty list",
			setupMock: func(mas *MockActorService) {
				mas.On("GetAll").Return([]domain.Actor{}, nil)
			},
			expectedResult: dto.ActorsListResponse{
				Actors: []dto.ActorResponse{},
			},
			expectedError: false,
		},
		{
			name: "service error",
			setupMock: func(mas *MockActorService) {
				mas.On("GetAll").Return([]domain.Actor{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			result, err := controller.ListActors(&gin.Context{})

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

func TestActorController_GetAllActorsWithMovies(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockActorService)
		expectedResult dto.ActorsWithFilmsListResponse
		expectedError  bool
	}{
		{
			name: "success",
			setupMock: func(mas *MockActorService) {
				mas.On("GetAllActorsWithMovies").Return([]domain.Actor{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
						Movies: []domain.Movie{
							{
								ID:          1,
								Title:       "Movie 1",
								Description: "Description 1",
								ReleaseYear: 2020,
								Rating:      8.5,
							},
						},
					},
				}, nil)
			},
			expectedResult: dto.ActorsWithFilmsListResponse{
				Actors: []dto.ActorWithFilms{
					{
						ID:        1,
						Name:      "Actor 1",
						Gender:    "male",
						BirthDate: "1990-01-01",
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
				},
			},
			expectedError: false,
		},
		{
			name: "service error",
			setupMock: func(mas *MockActorService) {
				mas.On("GetAllActorsWithMovies").Return([]domain.Actor{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockActorService{}
			tt.setupMock(mockService)

			controller := NewActorController(mockService)

			result, err := controller.GetAllActorsWithMovies(&gin.Context{})

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
