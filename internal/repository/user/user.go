package user

import (
	"context"
	"fmt"

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

func (u *User) Create(ctx context.Context, user entity.User) error {
	query := `INSERT INTO users (login, hash) VALUES($1, $2)`
	tag, err := u.pool.Exec(ctx, query, user.Login, user.Hash)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", errors.Trasform(err))
	}

	rowsAffectedCount := tag.RowsAffected()
	if rowsAffectedCount != 1 {
		return fmt.Errorf("expected one row to be affected, actually affected %d", rowsAffectedCount)
	}

	return nil
}
