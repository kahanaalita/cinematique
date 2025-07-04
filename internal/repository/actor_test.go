package repository

import (
	"cinematique/internal/domain"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestActorRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)
	birthDate, _ := time.Parse("2006-01-02", "1980-01-01")

	tests := []struct {
		name    string
		actor   domain.Actor
		setup   func()
		wantID  int
		wantErr bool
	}{
		{
			name: "successful actor creation",
			actor: domain.Actor{
				Name:      "Leonardo DiCaprio",
				Gender:    "male",
				BirthDate: birthDate,
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO actors \(name,gender,birth_date\) VALUES \(\$1,\$2,\$3\) RETURNING id`).
					WithArgs("Leonardo DiCaprio", "male", birthDate).
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

			gotID, err := repo.Create(tt.actor)

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

func TestActorRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)
	birthDate, _ := time.Parse("2006-01-02", "1980-01-01")

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    domain.Actor
		wantErr bool
	}{
		{
			name: "actor found",
			id:   1,
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"}).
					AddRow(1, "Leonardo DiCaprio", "male", birthDate)
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: domain.Actor{
				ID:        1,
				Name:      "Leonardo DiCaprio",
				Gender:    "male",
				BirthDate: birthDate,
			},
		},
		{
			name: "actor not found",
			id:   999,
			setup: func() {
				mock.ExpectQuery(`^SELECT`).
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

func TestActorRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)
	birthDate, _ := time.Parse("2006-01-02", "1980-01-01")

	tests := []struct {
		name    string
		actor   domain.Actor
		setup   func()
		wantErr bool
	}{
		{
			name: "successful update",
			actor: domain.Actor{
				ID:        1,
				Name:      "Leonardo DiCaprio Updated",
				Gender:    "male",
				BirthDate: birthDate,
			},
			setup: func() {
				mock.ExpectExec(`UPDATE actors SET name = \$1, gender = \$2, birth_date = \$3 WHERE id = \$4`).
					WithArgs("Leonardo DiCaprio Updated", "male", birthDate, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "actor not found",
			actor: domain.Actor{
				ID: 999,
			},
			setup: func() {
				mock.ExpectExec(`UPDATE actors SET name = \$1, gender = \$2, birth_date = \$3 WHERE id = \$4`).
					WithArgs("", "", time.Time{}, 999).
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

			err := repo.Update(tt.actor)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestActorRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)

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
				// Мок для проверки существования актёра
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"}).
						AddRow(1, "Test Actor", "male", time.Now()))

				mock.ExpectBegin()
				mock.ExpectExec(`^DELETE FROM film_actor WHERE actor_id = \$1$`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`^DELETE FROM actors WHERE id = \$1$`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "actor not found",
			id:   999,
			setup: func() {
				// Мок для проверки несуществующего актёра
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
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

func TestActorRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)
	birthDate1, _ := time.Parse("2006-01-02", "1980-01-01")
	birthDate2, _ := time.Parse("2006-01-02", "1985-05-15")

	tests := []struct {
		name    string
		setup   func()
		want    []domain.Actor
		wantErr bool
	}{
		{
			name: "get all actors",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"}).
					AddRow(1, "Leonardo DiCaprio", "male", birthDate1).
					AddRow(2, "Scarlett Johansson", "female", birthDate2)
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors$`).
					WillReturnRows(rows)
			},
			want: []domain.Actor{
				{
					ID:        1,
					Name:      "Leonardo DiCaprio",
					Gender:    "male",
					BirthDate: birthDate1,
				},
				{
					ID:        2,
					Name:      "Scarlett Johansson",
					Gender:    "female",
					BirthDate: birthDate2,
				},
			},
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
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestActorRepository_PartialUpdateActor(t *testing.T) {
	newName := "Brad Pitt"
	birthDate, _ := time.Parse("2006-01-02", "1980-01-01")

	tests := []struct {
		name    string
		id      int
		update  domain.ActorUpdate
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "partial update name",
			id:     1,
			update: domain.ActorUpdate{Name: &newName},
			setup: func(mock sqlmock.Sqlmock) {
				// First expect the actor existence check
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "gender", "birth_date"}).AddRow(1, "Old Name", "male", birthDate))

				// Then expect the column existence check with a flexible regex pattern
				expectedSQL := `SELECT EXISTS \(\s*SELECT 1\s+FROM information_schema\.columns\s+WHERE table_name = \$1 AND column_name = \$2\s*\)`
				mock.ExpectQuery(expectedSQL).
					WithArgs("actors", "updated_at").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				// Finally expect the update query (squirrel adds a space at the end)
				mock.ExpectExec(`^UPDATE actors SET name = \$1 WHERE id = \$2$`).
					WithArgs(newName, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "actor not found",
			id:     999,
			update: domain.ActorUpdate{Name: &newName},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:   "db error on actor check",
			id:     1,
			update: domain.ActorUpdate{Name: &newName},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT id, name, gender, birth_date FROM actors WHERE id = \$1$`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock for each test case to ensure clean state
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewActor(db)
			if tt.setup != nil {
				tt.setup(mock)
			}

			err = repo.PartialUpdateActor(tt.id, tt.update)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestActorRepository_GetMovies(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)

	tests := []struct {
		name    string
		actorID int
		setup   func()
		want    []domain.Movie
		wantErr bool
	}{
		{
			name:    "get movies for actor",
			actorID: 1,
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}).
					AddRow(1, "Inception", "A thief who steals corporate secrets...", 2010, 8.8).
					AddRow(2, "The Revenant", "A frontiersman on a fur trading...", 2015, 8.0)

				mock.ExpectQuery(`^SELECT f\.id, f\.title, f\.description, f\.release_year, f\.rating FROM films f JOIN film_actor fa ON f\.id = fa\.film_id WHERE fa\.actor_id = \$1$`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: []domain.Movie{
				{
					ID:          1,
					Title:       "Inception",
					Description: "A thief who steals corporate secrets...",
					ReleaseYear: 2010,
					Rating:      8.8,
				},
				{
					ID:          2,
					Title:       "The Revenant",
					Description: "A frontiersman on a fur trading...",
					ReleaseYear: 2015,
					Rating:      8.0,
				},
			},
		},
		{
			name:    "no movies for actor",
			actorID: 2,
			setup: func() {
				mock.ExpectQuery(`^SELECT`).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "release_year", "rating"}))
			},
			want: []domain.Movie{},
		},
		{
			name:    "database error",
			actorID: 1,
			setup: func() {
				mock.ExpectQuery(`^SELECT`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := repo.GetMovies(tt.actorID)

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

func TestActorRepository_GetAllActorsWithMovies(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewActor(db)
	birthDate1, _ := time.Parse("2006-01-02", "1980-01-01")
	birthDate2, _ := time.Parse("2006-01-02", "1985-05-15")

	tests := []struct {
		name    string
		setup   func()
		want    []domain.Actor
		wantErr bool
	}{
		{
			name: "get all actors with movies",
			setup: func() {
				rows := sqlmock.NewRows([]string{
					"a.id", "a.name", "a.gender", "a.birth_date",
					"f.id", "f.title", "f.description", "f.release_year", "f.rating",
				}).
					AddRow(1, "Leonardo DiCaprio", "male", birthDate1, 1, "Inception", "A thief...", 2010, 8.8).
					AddRow(1, "Leonardo DiCaprio", "male", birthDate1, 2, "The Revenant", "A frontiersman...", 2015, 8.0).
					AddRow(2, "Scarlett Johansson", "female", birthDate2, 3, "Lost in Translation", "A faded movie star...", 2003, 7.7)

				mock.ExpectQuery(`^SELECT a\.id, a\.name, a\.gender, a\.birth_date, f\.id, f\.title, f\.description, f\.release_year, f\.rating FROM actors a LEFT JOIN film_actor fa ON a\.id = fa\.actor_id LEFT JOIN films f ON fa\.film_id = f\.id ORDER BY a\.id, f\.id$`).
					WillReturnRows(rows)
			},
			want: []domain.Actor{
				{
					ID:        1,
					Name:      "Leonardo DiCaprio",
					Gender:    "male",
					BirthDate: birthDate1,
					Movies: []domain.Movie{
						{
							ID:          1,
							Title:       "Inception",
							Description: "A thief...",
							ReleaseYear: 2010,
							Rating:      8.8,
						},
						{
							ID:          2,
							Title:       "The Revenant",
							Description: "A frontiersman...",
							ReleaseYear: 2015,
							Rating:      8.0,
						},
					},
				},
				{
					ID:        2,
					Name:      "Scarlett Johansson",
					Gender:    "female",
					BirthDate: birthDate2,
					Movies: []domain.Movie{
						{
							ID:          3,
							Title:       "Lost in Translation",
							Description: "A faded movie star...",
							ReleaseYear: 2003,
							Rating:      7.7,
						},
					},
				},
			},
		},
		{
			name: "database error",
			setup: func() {
				mock.ExpectQuery(`^SELECT`).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := repo.GetAllActorsWithMovies()

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
