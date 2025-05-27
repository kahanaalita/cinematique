package controller

import (
	"github.com/gin-gonic/gin"

	"cinematigue/internal/controller/dto"
	"cinematigue/internal/domain"
)

type movieController struct {
	movieService ServiceMovie
}

func NewMovieController(movieService ServiceMovie) *movieController {
	return &movieController{
		movieService: movieService,
	}
}

func (c *movieController) CreateMovie(ctx *gin.Context, req dto.CreateMovieRequest) (dto.MovieResponse, error) {
	movie := domain.Movie{
		Title:       req.Title,
		Description: req.Description,
		ReleaseYear: req.ReleaseYear,
		Rating:      req.Rating,
	}

	id, err := c.movieService.Create(movie)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	return dto.MovieResponse{
		ID:          id,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Rating:      movie.Rating,
	}, nil
}

func (c *movieController) GetMovieByID(ctx *gin.Context, id int) (dto.MovieResponse, error) {
	movie, err := c.movieService.GetByID(id)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	return dto.MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Rating:      movie.Rating,
	}, nil
}

func (c *movieController) UpdateMovie(ctx *gin.Context, id int, req dto.UpdateMovieRequest) (dto.MovieResponse, error) {
	movie, err := c.movieService.GetByID(id)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	// Обновляем только переданные поля
	if req.Title != nil {
		movie.Title = *req.Title
	}
	if req.Description != nil {
		movie.Description = *req.Description
	}
	if req.ReleaseYear != nil {
		movie.ReleaseYear = *req.ReleaseYear
	}
	if req.Rating != nil {
		movie.Rating = *req.Rating
	}

	err = c.movieService.Update(movie)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	return dto.MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Rating:      movie.Rating,
	}, nil
}

func (c *movieController) DeleteMovie(ctx *gin.Context, id int) error {
	return c.movieService.Delete(id)
}

func (c *movieController) ListMovies(ctx *gin.Context) (dto.MoviesListResponse, error) {
	movies, err := c.movieService.GetAll()
	if err != nil {
		return dto.MoviesListResponse{}, err
	}

	response := dto.MoviesListResponse{
		Movies: make([]dto.MovieResponse, 0, len(movies)),
	}

	for _, movie := range movies {
		response.Movies = append(response.Movies, dto.MovieResponse{
			ID:          movie.ID,
			Title:       movie.Title,
			Description: movie.Description,
			ReleaseYear: movie.ReleaseYear,
			Rating:      movie.Rating,
		})
	}

	return response, nil
}
