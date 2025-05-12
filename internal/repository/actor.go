package repository

import (
	"database/sql"
	"errors"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// Actor представляет собой структуру актера из БД
type Actor struct {
	ID        int
	Name      string
	Gender    string
	BirthDate time.Time
}

// actor - репозиторий для работы с актерами
type actor struct {
	db *sql.DB
}

// NewActor создает новый репозиторий для работы с актерами
func NewActor(db *sql.DB) *actor {
	return &actor{db: db}
}

// Create создает нового актера в базе данных
func (a *actor) Create(actor Actor) (int, error) {
	query, args, err := sq.Insert("actors").
		Columns("name", "gender", "birth_date").
		Values(actor.Name, actor.Gender, actor.BirthDate).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}
	var id int
	err = a.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		log.Printf("Error creating actor: %v", err)
		return 0, err
	}
	return id, nil
}

// GetByID получает актера по ID
func (a *actor) GetByID(id int) (Actor, error) {
	query, args, err := sq.Select("id", "name", "gender", "birth_date").
		From("actors").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return Actor{}, err
	}
	var actor Actor
	err = a.db.QueryRow(query, args...).Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.BirthDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Actor{}, errors.New("actor not found")
		}
		return Actor{}, err
	}
	return actor, nil
}

// Update обновляет информацию об актере
func (a *actor) Update(actor Actor) error {
	query, args, err := sq.Update("actors").
		Set("name", actor.Name).
		Set("gender", actor.Gender).
		Set("birth_date", actor.BirthDate).
		Where(sq.Eq{"id": actor.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = a.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating actor: %v", err)
		return err
	}
	return nil
}

// Delete удаляет актера по ID
func (a *actor) Delete(id int) error {
	query, args, err := sq.Delete("actors").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = a.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error deleting actor: %v", err)
		return err
	}
	return nil
}

// GetAll возвращает всех актеров
func (a *actor) GetAll() ([]Actor, error) {
	query, args, err := sq.Select("id", "name", "gender", "birth_date").
		From("actors").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	actors := make([]Actor, 0)
	for rows.Next() {
		var actor Actor
		if err := rows.Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.BirthDate); err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actors, nil
}

// GetMovies возвращает все фильмы, в которых снимался актер
func (a *actor) GetMovies(actorID int) ([]Movie, error) {
	query, args, err := sq.Select("f.id", "f.title", "f.description", "f.release_year", "f.rating").
		From("films f").
		Join("film_actor fa ON f.id = fa.film_id").
		Where(sq.Eq{"fa.actor_id": actorID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := a.db.Query(query, args...)
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
