package repository

import (
	"cinematigue/internal/domain"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"log"
)

// movie представляет репозиторий фильмов.
type movie struct {
	db *sql.DB // соединение с базой данных
}

// NewMovie создаёт новый репозиторий фильмов.
func NewMovie(db *sql.DB) *movie {
	return &movie{db: db}
}

// Create создаёт новый фильм в базе данных.
func (m *movie) Create(movie domain.Movie) (int, error) {
	// реализация создания фильма
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

// GetByID возвращает фильм по заданному ID.
func (m *movie) GetByID(id int) (domain.Movie, error) {
	// реализация получения фильма по ID
	query, args, err := sq.Select("id", "title", "description", "release_year", "rating").
		From("films").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return domain.Movie{}, err
	}
	var movie domain.Movie
	err = m.db.QueryRow(query, args...).Scan(&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Rating)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movie{}, errors.New("movie not found")
		}
		return domain.Movie{}, err
	}
	return movie, nil
}

// Update обновляет информацию о фильме.
func (m *movie) Update(movie domain.Movie) error {
	// реализация обновления фильма
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

// Delete удаляет фильм по заданному ID.
func (m *movie) Delete(id int) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем связи с актёрами
	delFilmActor, args, err := sq.Delete("film_actor").
		Where(sq.Eq{"film_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete film_actor query: %w", err)
	}

	if _, err = tx.Exec(delFilmActor, args...); err != nil {
		log.Printf("Error deleting film_actor relations: %v", err)
		return fmt.Errorf("failed to delete film_actor relations: %w", err)
	}

	// Удаляем фильм
	delFilm, args, err := sq.Delete("films").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete film query: %w", err)
	}

	if _, err = tx.Exec(delFilm, args...); err != nil {
		log.Printf("Error deleting film: %v", err)
		return fmt.Errorf("failed to delete film: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAll возвращает все фильмы.
func (m *movie) GetAll() ([]domain.Movie, error) {
	// реализация получения всех фильмов
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
	movies := make([]domain.Movie, 0)
	for rows.Next() {
		var movie domain.Movie
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

// AddActor добавляет актёра к фильму.
func (m *movie) AddActor(movieID, actorID int) error {
	query, args, err := sq.Insert("film_actor").
		Columns("film_id", "actor_id").
		Values(movieID, actorID).
		Suffix("ON CONFLICT DO NOTHING").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build add actor query: %w", err)
	}

	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error adding actor to movie: %v", err)
		return fmt.Errorf("failed to add actor to movie: %w", err)
	}
	return nil
}

// RemoveActor удаляет актёра из фильма.
func (m *movie) RemoveActor(movieID, actorID int) error {
	// реализация удаления актёра из фильма
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

// GetActorsForMovieByID возвращает актёров фильма.
func (m *movie) GetActorsForMovieByID(movieID int) ([]domain.Actor, error) {
	// реализация получения актёров фильма
	query, args, err := sq.Select("a.id", "a.name", "a.gender", "a.birth_date").
		From("actors a").
		Join("film_actor fa ON a.id = fa.actor_id").
		Where(sq.Eq{"fa.film_id": movieID}).
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

	var actors []domain.Actor
	for rows.Next() {
		var actor domain.Actor
		err := rows.Scan(
			&actor.ID,
			&actor.Name,
			&actor.Gender,
			&actor.BirthDate,
		)
		if err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return actors, nil
}

// RemoveAllActors удаляет всех актёров из фильма.
func (m *movie) RemoveAllActors(movieID int) error {
	// реализация удаления всех актёров из фильма
	query, args, err := sq.Delete("film_actor").
		Where(sq.Eq{"film_id": movieID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error removing all actors from movie: %v", err)
		return err
	}
	return nil
}

// CreateMovieWithActors создаёт фильм с актёрами.
func (m *movie) CreateMovieWithActors(movie domain.Movie, actorIDs []int) (int, error) {
	tx, err := m.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаём фильм
	query, args, err := sq.Insert("films").
		Columns("title", "description", "release_year", "rating").
		Values(movie.Title, movie.Description, movie.ReleaseYear, movie.Rating).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build create movie query: %w", err)
	}

	var movieID int
	err = tx.QueryRow(query, args...).Scan(&movieID)
	if err != nil {
		log.Printf("Error creating movie: %v", err)
		return 0, fmt.Errorf("failed to create movie: %w", err)
	}

	// Добавляем связи с актёрами
	if len(actorIDs) > 0 {
		insertBuilder := sq.Insert("film_actor").Columns("film_id", "actor_id")
		for _, actorID := range actorIDs {
			insertBuilder = insertBuilder.Values(movieID, actorID)
		}

		query, args, err = insertBuilder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return 0, fmt.Errorf("failed to build add actors query: %w", err)
		}

		if _, err = tx.Exec(query, args...); err != nil {
			log.Printf("Error adding actors to movie: %v", err)
			return 0, fmt.Errorf("failed to add actors to movie: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return movieID, nil
}

// UpdateMovieActors обновляет актёров фильма.
func (m *movie) UpdateMovieActors(movieID int, actorIDs []int) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем все существующие связи фильма
	delQuery, delArgs, err := sq.Delete("film_actor").
		Where(sq.Eq{"film_id": movieID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete film_actor query: %w", err)
	}

	if _, err = tx.Exec(delQuery, delArgs...); err != nil {
		log.Printf("Error deleting film_actor relations: %v", err)
		return fmt.Errorf("failed to delete film_actor relations: %w", err)
	}

	// Добавляем новые связи, если они есть
	if len(actorIDs) > 0 {
		insertBuilder := sq.Insert("film_actor").Columns("film_id", "actor_id")
		for _, actorID := range actorIDs {
			insertBuilder = insertBuilder.Values(movieID, actorID)
		}

		insertQuery, insertArgs, err := insertBuilder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("failed to build insert film_actor query: %w", err)
		}

		if _, err = tx.Exec(insertQuery, insertArgs...); err != nil {
			log.Printf("Error adding actors to movie: %v", err)
			return fmt.Errorf("failed to add actors to movie: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMoviesForActor возвращает фильмы по актёру.
func (m *movie) GetMoviesForActor(actorID int) ([]domain.Movie, error) {
	// реализация получения фильмов по актёру
	query, args, err := sq.Select("f.id", "f.title", "f.description", "f.release_year", "f.rating").
		From("films f").
		Join("film_actor fa ON f.id = fa.film_id").
		Where(sq.Eq{"fa.actor_id": actorID}).
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

	movies := make([]domain.Movie, 0)
	for rows.Next() {
		var movie domain.Movie
		if err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.ReleaseYear,
			&movie.Rating,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

// SearchMoviesByTitle ищет фильмы по названию.
func (m *movie) SearchMoviesByTitle(titleFragment string) ([]domain.Movie, error) {
	// реализация поиска фильмов по названию
	query, args, err := sq.Select("id", "title", "description", "release_year", "rating").
		From("films").
		Where("title ILIKE $1", "%"+titleFragment+"%"). // PostgreSQL ILIKE для case-insensitive поиска
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
	var movies []domain.Movie
	for rows.Next() {
		var movie domain.Movie
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

// SearchMoviesByActorName ищет фильмы по имени актёра.
func (m *movie) SearchMoviesByActorName(actorNameFragment string) ([]domain.Movie, error) {
	// реализация поиска фильмов по имени актёра
	query, args, err := sq.Select("f.id", "f.title", "f.description", "f.release_year", "f.rating").
		From("films f").
		Join("film_actor fa ON f.id = fa.film_id").
		Join("actors a ON fa.actor_id = a.id").
		Where("a.name ILIKE $1", "%"+actorNameFragment+"%").
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
	var movies []domain.Movie
	for rows.Next() {
		var movie domain.Movie
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

// GetAllMoviesSorted возвращает фильмы с сортировкой.
func (m *movie) GetAllMoviesSorted(sortField, sortOrder string) ([]domain.Movie, error) {
	// реализация получения фильмов с сортировкой
	// Валидация поля сортировки
	allowedFields := map[string]bool{"title": true, "rating": true, "release_year": true}
	if !allowedFields[sortField] {
		sortField = "rating"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query := sq.Select("id", "title", "description", "release_year", "rating").
		From("films").
		OrderBy(sortField + " " + sortOrder).
		PlaceholderFormat(sq.Dollar)
	qstr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := m.db.Query(qstr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var movies []domain.Movie
	for rows.Next() {
		var movie domain.Movie
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

// PartialUpdateMovie частично обновляет фильм.
func (m *movie) PartialUpdateMovie(id int, update domain.MovieUpdate) error {
	// реализация частичного обновления фильма
	builder := sq.Update("films").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)
	if update.Title != nil {
		builder = builder.Set("title", *update.Title)
	}
	if update.Description != nil {
		builder = builder.Set("description", *update.Description)
	}
	if update.ReleaseYear != nil {
		builder = builder.Set("release_year", *update.ReleaseYear)
	}
	if update.Rating != nil {
		builder = builder.Set("rating", *update.Rating)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error partial updating movie: %v", err)
		return err
	}
	return nil
}
