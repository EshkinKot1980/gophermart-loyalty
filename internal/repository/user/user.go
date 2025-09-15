package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
)

type User struct {
	pool *pgxpool.Pool
}

func New(p *pgxpool.Pool) *User {
	return &User{pool: p}
}

func (u *User) GetByID(ctx context.Context, id uint64) (entity.User, error) {
	var user entity.User
	query := `SELECT id, login, hash, created_at FROM users WHERE id = $1`
	row := u.pool.QueryRow(ctx, query, id)

	err := row.Scan(&user.ID, &user.Login, &user.Hash, &user.Created)
	if err != nil {
		return entity.User{}, errors.Trasform(err)
	}

	return user, nil
}

func (u *User) Create(ctx context.Context, user entity.User) (entity.User, error) {
	query := `INSERT INTO users (login, hash) VALUES($1, $2) RETURNING id, created_at`
	row := u.pool.QueryRow(ctx, query, user.Login, user.Hash)

	err := row.Scan(&user.ID, &user.Created)
	if err != nil {
		return user, errors.Trasform(err)
	}

	return user, nil
}

func (u *User) FindByLogin(ctx context.Context, login string) (entity.User, error) {
	var user entity.User
	query := `SELECT id, login, hash, created_at FROM users WHERE login = $1`
	row := u.pool.QueryRow(ctx, query, login)

	err := row.Scan(&user.ID, &user.Login, &user.Hash, &user.Created)
	if err != nil {
		return entity.User{}, errors.Trasform(err)
	}

	return user, nil
}
