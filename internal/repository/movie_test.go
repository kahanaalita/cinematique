package repository

import (
	"cinematique/internal/domain"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"time"
)

func TestMovieRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)

	tests := []struct {
		name    string
		movie   domain.Movie
		setup   func()
		wantID  int
		wantErr bool
	}{
		{
			name: "successful movie creation",
			movie: domain.Movie{
				Title:       "Inception",
				Description: "A mind-bending movie",
				ReleaseYear: 2010,
				Rating:      8.8,
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO films \(title,description,release_year,rating\) VALUES \(\$1,\$2,\$3,\$4\) RETURNING id`).
					WithArgs("Inception", "A mind-bending movie", 2010, 8.8).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			gotID, err := repo.Create(tt.movie)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    domain.Movie
		wantErr bool
	}{
		{
			name: "movie found",
			id:   1,
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A mind-bending movie", 2010, 8.8)
				mock.ExpectQuery(`SELECT.* FROM films WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: domain.Movie{
				ID:          1,
				Title:       "Inception",
				Description: "A mind-bending movie",
				ReleaseYear: 2010,
				Rating:      8.8,
			},
		},
		{
			name: "movie not found",
			id:   999,
			setup: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := repo.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)

	tests := []struct {
		name    string
		movie   domain.Movie
		setup   func()
		wantErr bool
	}{
		{
			name: "successful update",
			movie: domain.Movie{
				ID:          1,
				Title:       "Inception Updated",
				Description: "Updated description",
				ReleaseYear: 2011,
				Rating:      9.0,
			},
			setup: func() {
				mock.ExpectExec(`UPDATE films SET title = \$1, description = \$2, release_year = \$3, rating = \$4 WHERE id = \$5`).
					WithArgs("Inception Updated", "Updated description", 2011, 9.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "movie not found",
			movie: domain.Movie{
				ID: 999,
			},
			setup: func() {
				mock.ExpectExec(`UPDATE films SET .*`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := repo.Update(tt.movie)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)

	tests := []struct {
		name    string
		id      int
		setup   func()
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   1,
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM film_actor WHERE film_id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`DELETE FROM films WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "movie not found",
			id:   999,
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM film_actor WHERE film_id = \$1`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := repo.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)

	tests := []struct {
		name    string
		setup   func()
		want    []domain.Movie
		wantErr bool
	}{
		{
			name: "get all movies",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A mind-bending movie", 2010, 8.8).
					AddRow(2, "The Revenant", "A survival story", 2015, 8.0)
				mock.ExpectQuery(`SELECT id, title, description, release_year, rating FROM films`).WillReturnRows(rows)
			},
			want: []domain.Movie{
				{ID: 1, Title: "Inception", Description: "A mind-bending movie", ReleaseYear: 2010, Rating: 8.8},
				{ID: 2, Title: "The Revenant", Description: "A survival story", ReleaseYear: 2015, Rating: 8.0},
			},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name: "no movies",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"})
				mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
			},
			want: []domain.Movie{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.GetAll()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_AddActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	tests := []struct {
		name    string
		movieID int
		actorID int
		setup   func()
		wantErr bool
	}{
		{
			name:    "add actor to movie",
			movieID: 1,
			actorID: 2,
			setup: func() {
				mock.ExpectExec(`INSERT INTO film_actor`).WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "db error",
			movieID: 1,
			actorID: 2,
			setup: func() {
				mock.ExpectExec(`INSERT INTO film_actor`).WithArgs(1, 2).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := repo.AddActor(tt.movieID, tt.actorID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_RemoveActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	movieID := 1
	actorID := 2
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "remove actor from movie",
			setup: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE actor_id = $1 AND film_id = $2")).WithArgs(actorID, movieID).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE actor_id = $1 AND film_id = $2")).WithArgs(actorID, movieID).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := repo.RemoveActor(movieID, actorID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_GetActorsForMovieByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	tests := []struct {
		name    string
		movieID int
		setup   func()
		want    []domain.Actor
		wantErr bool
	}{
		{
			name:    "get actors for movie",
			movieID: 1,
			setup: func() {
				// birth_date: используем корректный sql.NullTime
				birth := time.Date(1974, 11, 11, 0, 0, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"}).
					AddRow(1, "Leonardo DiCaprio", "male", sql.NullTime{Time: birth, Valid: true})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.name, a.gender, a.birth_date FROM actors a JOIN film_actor fa ON a.id = fa.actor_id WHERE fa.film_id = $1")).WithArgs(1).WillReturnRows(rows)
			},
			want: []domain.Actor{{ID: 1, Name: "Leonardo DiCaprio", Gender: "male", BirthDate: time.Date(1974, 11, 11, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:    "no actors",
			movieID: 2,
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.name, a.gender, a.birth_date FROM actors a JOIN film_actor fa ON a.id = fa.actor_id WHERE fa.film_id = $1")).WithArgs(2).WillReturnRows(rows)
			},
			want: []domain.Actor{},
		},
		{
			name:    "db error",
			movieID: 3,
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.name, a.gender, a.birth_date FROM actors a JOIN film_actor fa ON a.id = fa.actor_id WHERE fa.film_id = $1")).WithArgs(3).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.GetActorsForMovieByID(tt.movieID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_RemoveAllActors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	tests := []struct {
		name    string
		movieID int
		setup   func()
		wantErr bool
	}{
		{
			name:    "remove all actors",
			movieID: 1,
			setup: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE film_id = $1")).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 2))
			},
		},
		{
			name:    "db error",
			movieID: 2,
			setup: func() {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE film_id = $1")).WithArgs(2).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := repo.RemoveAllActors(tt.movieID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_CreateMovieWithActors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	tests := []struct {
		name    string
		setup   func()
		wantID  int
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO films (title,description,release_year,rating) VALUES ($1,$2,$3,$4) RETURNING id")).
					WithArgs("Test Movie", "desc", 2020, 7.5).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO film_actor (film_id,actor_id) VALUES ($1,$2)")).
					WithArgs(10, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantID: 10,
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO films (title,description,release_year,rating) VALUES ($1,$2,$3,$4) RETURNING id")).
					WithArgs("Test Movie", "desc", 2020, 7.5).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			id, err := repo.CreateMovieWithActors(domain.Movie{
				Title:       "Test Movie",
				Description: "desc",
				ReleaseYear: 2020,
				Rating:      7.5,
			}, []int{1})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_UpdateMovieActors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE film_id = $1")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO film_actor (film_id,actor_id) VALUES ($1,$2),($3,$4)")).
					WithArgs(1, 1, 1, 2).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM film_actor WHERE film_id = $1")).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := repo.UpdateMovieActors(1, []int{1, 2})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_GetMoviesForActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	actorID := 1
	tests := []struct {
		name    string
		setup   func()
		want    []domain.Movie
		wantErr bool
	}{
		{
			name: "get movies for actor",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A mind-bending movie", 2010, 8.8)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id WHERE fa.actor_id = $1")).WithArgs(actorID).WillReturnRows(rows)
			},
			want: []domain.Movie{{ID: 1, Title: "Inception", Description: "A mind-bending movie", ReleaseYear: 2010, Rating: 8.8}},
		},
		{
			name: "no movies",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id WHERE fa.actor_id = $1")).WithArgs(actorID).WillReturnRows(rows)
			},
			want: []domain.Movie{},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id WHERE fa.actor_id = $1")).WithArgs(actorID).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.GetMoviesForActor(actorID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_SearchMoviesByTitle(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	titleFragment := "incep"
	tests := []struct {
		name    string
		setup   func()
		want    []domain.Movie
		wantErr bool
	}{
		{
			name: "find movies by title",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A mind-bending movie", 2010, 8.8)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films WHERE title ILIKE $1")).WithArgs("%incep%").WillReturnRows(rows)
			},
			want: []domain.Movie{{ID: 1, Title: "Inception", Description: "A mind-bending movie", ReleaseYear: 2010, Rating: 8.8}},
		},
		{
			name: "no movies found",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films WHERE title ILIKE $1")).WithArgs("%incep%").WillReturnRows(rows)
			},
			want: []domain.Movie{},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films WHERE title ILIKE $1")).WithArgs("%incep%").WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.SearchMoviesByTitle(titleFragment)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_GetAllMoviesSorted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	sortField := "title"
	tests := []struct {
		name      string
		sortOrder string
		setup     func()
		want      []domain.Movie
		wantErr   bool
	}{
		{
			name:      "sorted movies ASC",
			sortOrder: "ASC",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "A", "desc", 2010, 7.1).
					AddRow(2, "B", "desc2", 2011, 8.1)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films ORDER BY " + sortField + " ASC")).WillReturnRows(rows)
			},
			want: []domain.Movie{
				{ID: 1, Title: "A", Description: "desc", ReleaseYear: 2010, Rating: 7.1},
				{ID: 2, Title: "B", Description: "desc2", ReleaseYear: 2011, Rating: 8.1},
			},
		},
		{
			name:      "sorted movies DESC",
			sortOrder: "DESC",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(2, "B", "desc2", 2011, 8.1).
					AddRow(1, "A", "desc", 2010, 7.1)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films ORDER BY " + sortField + " DESC")).WillReturnRows(rows)
			},
			want: []domain.Movie{
				{ID: 2, Title: "B", Description: "desc2", ReleaseYear: 2011, Rating: 8.1},
				{ID: 1, Title: "A", Description: "desc", ReleaseYear: 2010, Rating: 7.1},
			},
		},
		{
			name:      "db error",
			sortOrder: "ASC",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, release_year, rating FROM films ORDER BY " + sortField + " ASC")).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.GetAllMoviesSorted(sortField, tt.sortOrder)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_PartialUpdateMovie(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	title := "NewTitle"
	update := domain.MovieUpdate{Title: &title}
	id := 1
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "partial update success",
			setup: func() {
				mock.ExpectExec(`UPDATE films SET title = \$1 WHERE id = \$2`).WithArgs("NewTitle", id).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectExec(`UPDATE films SET title = \$1 WHERE id = \$2`).WithArgs("NewTitle", id).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := repo.PartialUpdateMovie(id, update)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMovieRepository_SearchMoviesByActorName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMovie(db)
	actorNameFragment := "leo"
	arg := "%leo%"
	tests := []struct {
		name    string
		setup   func()
		want    []domain.Movie
		wantErr bool
	}{
		{
			name: "find movies by actor name",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A mind-bending movie", 2010, 8.8)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id JOIN actors a ON fa.actor_id = a.id WHERE a.name ILIKE $1")).WithArgs(arg).WillReturnRows(rows)
			},
			want: []domain.Movie{{ID: 1, Title: "Inception", Description: "A mind-bending movie", ReleaseYear: 2010, Rating: 8.8}},
		},
		{
			name: "no movies found",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id JOIN actors a ON fa.actor_id = a.id WHERE a.name ILIKE $1")).WithArgs(arg).WillReturnRows(rows)
			},
			want: []domain.Movie{},
		},
		{
			name: "db error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id, f.title, f.description, f.release_year, f.rating FROM films f JOIN film_actor fa ON f.id = fa.film_id JOIN actors a ON fa.actor_id = a.id WHERE a.name ILIKE $1")).WithArgs(arg).WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.SearchMoviesByActorName(actorNameFragment)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
