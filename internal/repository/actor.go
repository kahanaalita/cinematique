package repository

import (
	"database/sql"
	"errors"
	"log"
	"time"
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
	var id int
	err := a.db.QueryRow(
		"INSERT INTO actors (name, gender, birth_date) VALUES ($1, $2, $3) RETURNING id",
		actor.Name, actor.Gender, actor.BirthDate,
	).Scan(&id)

	if err != nil {
		log.Printf("Error creating actor: %v", err)
		return 0, err
	}

	return id, nil
}

// GetByID получает актера по ID
func (a *actor) GetByID(id int) (Actor, error) {
	var actor Actor
	err := a.db.QueryRow(
		"SELECT id, name, gender, birth_date FROM actors WHERE id = $1",
		id,
	).Scan(&actor.ID, &actor.Name, &actor.Gender, &actor.BirthDate)

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

	_, err := a.db.Exec(
		"UPDATE actors SET name = $1, gender = $2, birth_date = $3 WHERE id = $4",
		actor.Name, actor.Gender, actor.BirthDate, actor.ID,
	)

	if err != nil {
		log.Printf("Error updating actor: %v", err)
		return err
	}

	return nil
}

// Delete удаляет актера по ID
func (a *actor) Delete(id int) error {
	_, err := a.db.Exec("DELETE FROM actors WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting actor: %v", err)
		return err
	}

	return nil
}

// GetAll возвращает всех актеров
func (a *actor) GetAll() ([]Actor, error) {
	rows, err := a.db.Query("SELECT id, name, gender, birth_date FROM actors")
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
	rows, err := a.db.Query(`
		SELECT f.id, f.title, f.description, f.release_year, f.rating
		FROM films f
		JOIN film_actor fa ON f.id = fa.film_id
		WHERE fa.actor_id = $1
	`, actorID)

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
