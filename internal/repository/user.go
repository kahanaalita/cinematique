package repository

import (
	"database/sql"
	"log"
	"time" // Добавляем импорт time

	"cinematique/internal/domain"
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
	start := time.Now()
	operation := "create_user"
	queryType := "INSERT"

	var id int
	query, args, err := sq.Insert("users").
		Columns("username", "email", "password_hash", "role").
		Values(user.Username, user.Email, user.PasswordHash, user.Role).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return 0, err
	}

	err = r.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return 0, err
	}
	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return id, nil
}

// GetByUsername возвращает пользователя по имени.
func (r *UserRepository) GetByUsername(username string) (domain.User, error) {
	start := time.Now()
	operation := "get_user_by_username"
	queryType := "SELECT"

	var user domain.User
	
	query, args, err := sq.Select("id", "username", "email", "password_hash", "role").
		From("users").
		Where(sq.Eq{"username": username}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.User{}, err
	}

	err = r.db.QueryRow(query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return domain.User{}, sql.ErrNoRows
		}
		log.Printf("Error getting user by username: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.User{}, err
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return user, nil
}

// GetByID возвращает пользователя по ID.
func (r *UserRepository) GetByID(id int) (domain.User, error) {
	start := time.Now()
	operation := "get_user_by_id"
	queryType := "SELECT"

	var user domain.User

	query, args, err := sq.Select("id", "username", "email", "password_hash", "role").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.User{}, err
	}

	err = r.db.QueryRow(query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
			return domain.User{}, sql.ErrNoRows
		}
		log.Printf("Error getting user by ID: %v", err)
		dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
		return domain.User{}, err
	}

	dbQueryDurationSeconds.WithLabelValues(operation, queryType).Observe(time.Since(start).Seconds())
	dbQueriesTotal.WithLabelValues(operation, queryType).Inc()
	return user, nil
}
