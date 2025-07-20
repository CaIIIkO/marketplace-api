package user

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create - сохраняет пользователя в БД
func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	query := `
		INSERT INTO users (login, login_lower, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, u.Login, strings.ToLower(u.Login), u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByEmail - возвращает пользователя по login_lower (или nil, если не найден)
func (r *Repository) GetByLogin(ctx context.Context, login string) (*User, error) {
	query := `
		SELECT id, login, password, created_at
		FROM users
		WHERE login_lower = $1
	`
	row := r.pool.QueryRow(ctx, query, strings.ToLower(login))

	var u User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
