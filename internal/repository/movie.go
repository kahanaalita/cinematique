package repository

import (
	"database/sql"
	"errors"
	"log"
)

// Movie представляет собой структуру фильма из БД
type Movie struct {
	ID          int
	Title       string
	Description string
	ReleaseYear int
	Rating      float64
}

// movie - репозиторий для работы с фильмами
type movie struct {
	db *sql.DB
}

// NewMovie создает новый репозиторий для работы с фильмами
func NewMovie(db *sql.DB) *movie {
	return &movie{db: db}
}

// Create создает новый фильм в базе данных
func (m *movie) Create(movie Movie) (int, error) {
	var id int
	err := m.db.QueryRow(
		"INSERT INTO films (title, description, release_year, rating) VALUES ($1, $2, $3, $4) RETURNING id",
		movie.Title, movie.Description, movie.ReleaseYear, movie.Rating,
	).Scan(&id)

	if err != nil {
		log.Printf("Error creating movie: %v", err)
		return 0, err
	}

	return id, nil
}

// GetByID получает фильм по ID
func (m *movie) GetByID(id int) (Movie, error) {
	var movie Movie
	err := m.db.QueryRow(
		"SELECT id, title, description, release_year, rating FROM films WHERE id = $1",
		id,
	).Scan(&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Rating)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Movie{}, errors.New("movie not found")
		}
		return Movie{}, err
	}

	return movie, nil
}

// Update обновляет информацию о фильме
func (m *movie) Update(movie Movie) error {
	_, err := m.db.Exec(
		"UPDATE films SET title = $1, description = $2, release_year = $3, rating = $4 WHERE id = $5",
		movie.Title, movie.Description, movie.ReleaseYear, movie.Rating, movie.ID,
	)

	if err != nil {
		log.Printf("Error updating movie: %v", err)
		return err
	}

	return nil
}

// Delete удаляет фильм по ID
func (m *movie) Delete(id int) error {
	_, err := m.db.Exec("DELETE FROM films WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting movie: %v", err)
		return err
	}

	return nil
}

// GetAll возвращает все фильмы
func (m *movie) GetAll() ([]Movie, error) {
	rows, err := m.db.Query("SELECT id, title, description, release_year, rating FROM films")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := make([]Movie, 0)
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Rating); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

// AddActor добавляет актера к фильму
func (m *movie) AddActor(movieID, actorID int) error {
	_, err := m.db.Exec(
		"INSERT INTO film_actor (film_id, actor_id) VALUES ($1, $2)",
		movieID, actorID,
	)

	if err != nil {
		log.Printf("Error adding actor to movie: %v", err)
		return err
	}

	return nil
}

// RemoveActor удаляет актера из фильма
func (m *movie) RemoveActor(movieID, actorID int) error {
	_, err := m.db.Exec(
		"DELETE FROM film_actor WHERE film_id = $1 AND actor_id = $2",
		movieID, actorID,
	)

	if err != nil {
		log.Printf("Error removing actor from movie: %v", err)
		return err
	}

	return nil
}
