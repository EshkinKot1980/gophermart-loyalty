package repository

import (
	"context"
	"fmt"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/pg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Withdrawals struct {
	pool *pgxpool.Pool
}

func NewWithdrawals(db *pg.DB) *Withdrawals {
	return &Withdrawals{pool: db.Pool()}
}

func (r *Withdrawals) Create(ctx context.Context, w entity.Withdrawals) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query :=
		`UPDATE balance 
			SET balance = balance - $2, debited = debited + $2
			WHERE user_id = $1 AND balance >= $2`

	tag, err := tx.Exec(ctx, query, w.UserID, w.Sum)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("failed to update balance: %w", errors.ErrNoRowsUpdated)
	}

	query = `INSERT INTO withdrawals (user_id, order_num, sum) VALUES($1, $2, $3)`
	_, err = tx.Exec(ctx, query, w.UserID, w.OrderNumber, w.Sum)
	if err != nil {
		return fmt.Errorf("failed to insert into withdrawals: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Withdrawals) GetAllByUser(ctx context.Context, userID uint64) ([]entity.Withdrawals, error) {
	var list []entity.Withdrawals
	query := `SELECT id, user_id, order_num, sum, processed_at FROM withdrawals WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return list, errors.Trasform(err)
	}

	list, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Withdrawals])
	if err != nil {
		return list, errors.Trasform(err)
	}

	return list, nil
}
