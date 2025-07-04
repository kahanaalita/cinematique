package repository

import (
	"cinematique/internal/domain"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"log"
	"time"
)

// actor реализует репозиторий для актёров
type actor struct {
	db *sql.DB // соединение с базой данных
}

// NewActor создаёт репозиторий актёров
func NewActor(db *sql.DB) *actor {
	return &actor{db: db}
}

// Create создаёт актёра
func (a *actor) Create(actor domain.Actor) (int, error) {
	start := time.Now()
	operation := "create_actor"
	queryType := "INSERT"

	query, args, err := sq.Insert("actors").
		Columns("name", "gender", "birth_date").
		Values(actor.Name, actor.Gender, actor.BirthDate).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return 0, err
	}
	var id int
	err = a.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		log.Printf("Error creating actor: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return 0, err
	}
	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return id, nil
}

// GetByID возвращает актёра по ID
func (a *actor) GetByID(id int) (domain.Actor, error) {
	start := time.Now()
	operation := "get_actor_by_id"
	queryType := "SELECT"

	query, args, err := sq.Select("id", "name", "gender", "birth_date").
		From("actors").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.Actor{}, fmt.Errorf("building query: %w", err)
	}

	var actor domain.Actor
	err = a.db.QueryRow(query, args...).Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.BirthDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return domain.Actor{}, domain.ErrActorNotFound
		}
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.Actor{}, fmt.Errorf("scanning actor: %w", err)
	}
	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return actor, nil
}

// Update обновляет актёра
func (a *actor) Update(actor domain.Actor) error {
	start := time.Now()
	operation := "update_actor"
	queryType := "UPDATE"

	query, args, err := sq.Update("actors").
		Set("name", actor.Name).
		Set("gender", actor.Gender).
		Set("birth_date", actor.BirthDate).
		Where(sq.Eq{"id": actor.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("building query: %w", err)
	}

	result, err := a.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating actor: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("executing update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.ErrActorNotFound
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return nil
}

// Delete удаляет актёра по ID
func (a *actor) Delete(id int) error {
	start := time.Now()
	operation := "delete_actor"
	queryType := "DELETE"

	// Сначала проверяем существование актёра
	_, err := a.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return domain.ErrActorNotFound
		}
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("checking actor existence: %w", err)
	}

	tx, err := a.db.Begin()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Сначала удаляем связи с фильмами
	delFilmActor, args, err := sq.Delete("film_actor").
		Where(sq.Eq{"actor_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to build delete film_actor query: %w", err)
	}

	if _, err = tx.Exec(delFilmActor, args...); err != nil {
		log.Printf("Error deleting film_actor relations: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to delete film_actor relations: %w", err)
	}

	// Затем удаляем самого актёра
	delActor, args, err := sq.Delete("actors").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to build delete actor query: %w", err)
	}

	if _, err = tx.Exec(delActor, args...); err != nil {
		log.Printf("Error deleting actor: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to delete actor: %w", err)
	}

	if err = tx.Commit(); err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return nil
}

// GetAll возвращает всех актёров
func (a *actor) GetAll() ([]domain.Actor, error) {
	start := time.Now()
	operation := "get_all_actors"
	queryType := "SELECT"

	query, args, err := sq.Select("id", "name", "gender", "birth_date").
		From("actors").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return nil, err
	}
	rows, err := a.db.Query(query, args...)
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return nil, err
	}
	defer rows.Close()
	var actors []domain.Actor
	for rows.Next() {
		var actor domain.Actor
		if err := rows.Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.BirthDate); err != nil {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return nil, err
		}
		actors = append(actors, actor)
	}
	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return actors, nil
}

// GetMovies возвращает фильмы актёра
func (a *actor) GetMovies(actorID int) ([]domain.Movie, error) {
	start := time.Now()
	operation := "get_movies_for_actor"
	queryType := "SELECT"

	query, args, err := sq.Select("f.id", "f.title", "f.description", "f.release_year", "f.rating").
		From("films f").
		Join("film_actor fa ON f.id = fa.film_id").
		Where(sq.Eq{"fa.actor_id": actorID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return []domain.Movie{}, err
	}
	rows, err := a.db.Query(query, args...)
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return []domain.Movie{}, err
	}
	defer rows.Close()
	movies := []domain.Movie{}
	for rows.Next() {
		var movie domain.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Rating); err != nil {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return []domain.Movie{}, err
		}
		movies = append(movies, movie)
	}
	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return movies, nil
}

// GetAllActorsWithMovies возвращает актёров с их фильмами
func (a *actor) GetAllActorsWithMovies() ([]domain.Actor, error) {
	start := time.Now()
	operation := "get_all_actors_with_movies"
	queryType := "SELECT"

	// Используем один запрос с JOIN вместо N+1 запросов
	query, args, err := sq.Select(
		"a.id", "a.name", "a.gender", "a.birth_date",
		"f.id", "f.title", "f.description", "f.release_year", "f.rating",
	).
		From("actors a").
		LeftJoin("film_actor fa ON a.id = fa.actor_id").
		LeftJoin("films f ON fa.film_id = f.id").
		OrderBy("a.id", "f.id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := a.db.Query(query, args...)
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var result []domain.Actor
	var currentActor *domain.Actor

	for rows.Next() {
		var (
			actorID        int
			actorName      string
			actorGender    string
			actorBirthDate time.Time
			movieID        sql.NullInt64
			movieTitle     sql.NullString
			movieDesc      sql.NullString
			releaseYear    sql.NullInt32
			rating         sql.NullFloat64
		)

		err = rows.Scan(
			&actorID, &actorName, &actorGender, &actorBirthDate,
			&movieID, &movieTitle, &movieDesc, &releaseYear, &rating,
		)
		if err != nil {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if currentActor == nil || currentActor.ID != actorID {
			// If we have a current actor, add it to the result before creating a new one
			if currentActor != nil {
				result = append(result, *currentActor)
			}
			currentActor = &domain.Actor{
				ID:        actorID,
				Name:      actorName,
				Gender:    actorGender,
				BirthDate: actorBirthDate,
				Movies:    []domain.Movie{},
			}
		}

		if movieID.Valid {
			currentActor.Movies = append(currentActor.Movies, domain.Movie{
				ID:          int(movieID.Int64),
				Title:       movieTitle.String,
				Description: movieDesc.String,
				ReleaseYear: int(releaseYear.Int32),
				Rating:      rating.Float64,
			})
		}
	}

	// Add the last actor to the result
	if currentActor != nil {
		result = append(result, *currentActor)
	}

	if err = rows.Err(); err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return result, nil
}

// PartialUpdateActor частично обновляет актёра
func (a *actor) PartialUpdateActor(id int, update domain.ActorUpdate) error {
	start := time.Now()
	operation := "partial_update_actor"
	queryType := "UPDATE"

	// Проверяем, что есть хотя бы одно поле для обновления
	if update.Name == nil && update.Gender == nil && update.BirthDate == nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("no fields to update")
	}

	// Проверяем существование актёра
	_, err := a.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return domain.ErrActorNotFound
		}
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("checking actor existence: %w", err)
	}

	// Строим запрос на обновление
	builder := sq.Update("actors").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	if update.Name != nil {
		builder = builder.Set("name", *update.Name)
	}
	if update.Gender != nil {
		builder = builder.Set("gender", *update.Gender)
	}
	if update.BirthDate != nil {
		builder = builder.Set("birth_date", *update.BirthDate)
	}

	// Добавляем updated_at, если поле существует в таблице
	hasUpdatedAt, err := a.columnExists("actors", "updated_at")
	if err != nil {
		log.Printf("Warning: failed to check updated_at column: %v", err)
	}
	if hasUpdatedAt {
		builder = builder.Set("updated_at", "NOW()")
	}

	query, args, err := builder.ToSql()
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := a.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error partially updating actor: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("failed to update actor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Warning: failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return fmt.Errorf("no rows were affected, actor with id %d may not exist", id)
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return nil
}

// columnExists проверяет существование колонки в таблице
func (a *actor) columnExists(tableName, columnName string) (bool, error) {
	start := time.Now()
	operation := "column_exists"
	queryType := "SELECT"

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		)`

	var exists bool
	err := a.db.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return false, fmt.Errorf("failed to check column existence: %w", err)
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return exists, nil
}