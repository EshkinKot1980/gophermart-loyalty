package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/pg"
)

type Order struct {
	pool *pgxpool.Pool
}

func NewOrder(db *pg.DB) *Order {
	return &Order{pool: db.Pool()}
}

func (w *Order) GetByNumber(ctx context.Context, number string) (order entity.Order, err error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE number = $1`

	rows, err := w.pool.Query(ctx, query, number)
	if err != nil {
		return order, fmt.Errorf("failed to select from orders: %w", err)
	}

	order, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return order, errors.Trasform(err)
	}

	return order, nil
}

func (w *Order) GetAllByUser(ctx context.Context, userID uint64) (orders []entity.Order, err error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE user_id = $1`

	rows, err := w.pool.Query(ctx, query, userID)
	if err != nil {
		return orders, fmt.Errorf("failed to select from orders: %w", err)
	}

	orders, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return orders, fmt.Errorf("failed to parse selected orders: %w", err)
	}

	return orders, nil
}

func (w *Order) Create(ctx context.Context, order entity.Order) error {
	query := `INSERT INTO orders (number, user_id, status) VALUES($1, $2, $3)`
	_, err := w.pool.Exec(ctx, query, order.Number, order.UserID, entity.OrderStatusNew)

	if err != nil {
		return errors.Trasform(err)
	}

	return nil
}
