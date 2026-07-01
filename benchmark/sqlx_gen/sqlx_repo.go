package sqlx_gen

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        int32     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

type Queries struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) GetUser(ctx context.Context, id int32) (User, error) {
	var user User
	err := q.db.GetContext(ctx, &user, "SELECT id, username, email, created_at FROM users WHERE id = $1 LIMIT 1", id)
	return user, err
}

func (q *Queries) ListUsers(ctx context.Context, limit, offset int32) ([]User, error) {
	var users []User
	err := q.db.SelectContext(ctx, &users, "SELECT id, username, email, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	return users, err
}

func (q *Queries) CreateUser(ctx context.Context, username, email string) (User, error) {
	var user User
	err := q.db.QueryRowxContext(ctx, "INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id, username, email, created_at", username, email).StructScan(&user)
	return user, err
}
