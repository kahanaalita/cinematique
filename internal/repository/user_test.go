package repository

import (
	"cinematigue/internal/domain"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		user    domain.User
		setup   func()
		want    int
		wantErr bool
	}{
		{
			name: "successful user creation",
			user: domain.User{
				Username:     "testuser",
				PasswordHash: "hashedpassword",
				Role:         "user",
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO users \(username,password_hash,role\) VALUES \(\$1,\$2,\$3\) RETURNING id`).
					WithArgs("testuser", "hashedpassword", "user").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			want: 1,
		},
		{
			name: "duplicate username",
			user: domain.User{
				Username:     "existinguser",
				PasswordHash: "hashedpassword",
				Role:         "user",
			},
			setup: func() {
				mock.ExpectQuery(`INSERT INTO users`).
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

			got, err := repo.CreateUser(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name     string
		username string
		setup    func()
		want     domain.User
		wantErr  bool
	}{
		{
			name:     "user found",
			username: "testuser",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "role"}).
					AddRow(1, "testuser", "hashedpass", "user")
				mock.ExpectQuery(`SELECT.* FROM users WHERE username = \$1`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			want: domain.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "hashedpass",
				Role:         "user",
			},
		},
		{
			name:     "user not found",
			username: "nonexistent",
			setup: func() {
				mock.ExpectQuery(`^SELECT`).
					WithArgs("nonexistent").
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

			got, err := repo.GetByUsername(tt.username)

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

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    domain.User
		wantErr bool
	}{
		{
			name: "user found",
			id:   1,
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "role"}).
					AddRow(1, "testuser", "hashedpass", "admin")
				mock.ExpectQuery(`^SELECT id, username, password_hash, role FROM users WHERE id = \$1$`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: domain.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "hashedpass",
				Role:         "admin",
			},
		},
		{
			name: "user not found",
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
