package repository

import (
	"database/sql"
	"errors"
	"log"

	sq "github.com/Masterminds/squirrel"
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
	query, args, err := sq.Insert("films").
		Columns("title", "description", "release_year", "rating").
		Values(movie.Title, movie.Description, movie.ReleaseYear, movie.Rating).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}
	var id int
	err = m.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		log.Printf("Error creating movie: %v", err)
		return 0, err
	}
	return id, nil
}

// GetByID получает фильм по ID
func (m *movie) GetByID(id int) (Movie, error) {
	query, args, err := sq.Select("id", "title", "description", "release_year", "rating").
		From("films").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return Movie{}, err
	}
	var movie Movie
	err = m.db.QueryRow(query, args...).Scan(&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Rating)
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
	query, args, err := sq.Update("films").
		Set("title", movie.Title).
		Set("description", movie.Description).
		Set("release_year", movie.ReleaseYear).
		Set("rating", movie.Rating).
		Where(sq.Eq{"id": movie.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating movie: %v", err)
		return err
	}
	return nil
}

// Delete удаляет фильм по ID
func (m *movie) Delete(id int) error {
	query, args, err := sq.Delete("films").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error deleting movie: %v", err)
		return err
	}
	return nil
}

// GetAll возвращает все фильмы
func (m *movie) GetAll() ([]Movie, error) {
	query, args, err := sq.Select("id", "title", "description", "release_year", "rating").
		From("films").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := m.db.Query(query, args...)
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
	query, args, err := sq.Insert("film_actor").
		Columns("film_id", "actor_id").
		Values(movieID, actorID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error adding actor to movie: %v", err)
		return err
	}
	return nil
}

// RemoveActor удаляет актера из фильма
func (m *movie) RemoveActor(movieID, actorID int) error {
	query, args, err := sq.Delete("film_actor").
		Where(sq.Eq{"film_id": movieID, "actor_id": actorID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error removing actor from movie: %v", err)
		return err
	}
	return nil
}
