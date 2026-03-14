package repository

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// User represents the user schema in the database.
type User struct {
	ID           int
	Username     string
	PasswordHash string
}

// UserRepository provides access to the users in the database.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(dbURL string) (*UserRepository, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to Postgres for Authentication")
	return &UserRepository{db: db}, nil
}

// Close closes the database connection.
func (r *UserRepository) Close() error {
	return r.db.Close()
}

// CreateUser inserts a new user into the database.
func (r *UserRepository) CreateUser(username, passwordHash string) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2)`
	_, err := r.db.Exec(query, username, passwordHash)
	return err
}

// GetUserByUsername retrieves a user by their username.
func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	var user User
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CountUsers returns the total number of users in the database.
func (r *UserRepository) CountUsers() (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
