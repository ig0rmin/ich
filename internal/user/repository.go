package user

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	var id int
	query := "INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := "SELECT id, email, username, password FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
