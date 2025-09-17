package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
)

type Balance struct {
	pool *pgxpool.Pool
}

func NewBalance(p *pgxpool.Pool) *Balance {
	return &Balance{pool: p}
}

func (b *Balance) GetByUser(ctx context.Context, userID uint64) (entity.Balance, error) {
	var balance entity.Balance
	query := `SELECT user_id, balance, debited FROM balance WHERE user_id = $1`
	row := b.pool.QueryRow(ctx, query, userID)

	err := row.Scan(&balance.UserID, &balance.Balance, &balance.Debited)
	if err != nil {
		return balance, errors.Trasform(err)
	}

	return balance, nil
}
