package repository

import (
	"database/sql"
	"log"

	"cinematigue/internal/domain"
	sq "github.com/Masterminds/squirrel"
)

// UserRepository реализует репозиторий пользователей.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создаёт нового пользователя.
func (r *UserRepository) CreateUser(user domain.User) (int, error) {
	var id int
	query, args, err := sq.Insert("users").
		Columns("username", "password_hash", "role").
		Values(user.Username, user.PasswordHash, user.Role).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return 0, err
	}

	err = r.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return 0, err
	}
	return id, nil
}

// GetByUsername возвращает пользователя по имени.
func (r *UserRepository) GetByUsername(username string) (domain.User, error) {
	var user domain.User
	
	query, args, err := sq.Select("id", "username", "password_hash", "role").
		From("users").
		Where(sq.Eq{"username": username}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return domain.User{}, err
	}

	err = r.db.QueryRow(query, args...).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, sql.ErrNoRows
		}
		log.Printf("Error getting user by username: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

// GetByID возвращает пользователя по ID.
func (r *UserRepository) GetByID(id int) (domain.User, error) {
	var user domain.User

	query, args, err := sq.Select("id", "username", "password_hash", "role").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return domain.User{}, err
	}

	err = r.db.QueryRow(query, args...).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, sql.ErrNoRows
		}
		log.Printf("Error getting user by ID: %v", err)
		return domain.User{}, err
	}

	return user, nil
}
