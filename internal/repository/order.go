package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
)

type Order struct {
	pool *pgxpool.Pool
}

func NewOrder(p *pgxpool.Pool) *Order {
	return &Order{pool: p}
}

func (o *Order) GetByNumber(ctx context.Context, number string) (order entity.Order, err error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE number = $1`

	rows, err := o.pool.Query(ctx, query, number)
	if err != nil {
		return order, errors.Trasform(err)
	}

	order, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return order, errors.Trasform(err)
	}

	return order, nil
}
func (o *Order) GetAllByUser(ctx context.Context, userID uint64) (orders []entity.Order, err error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE user_id = $1`

	rows, err := o.pool.Query(ctx, query, userID)
	if err != nil {
		return orders, errors.Trasform(err)
	}

	orders, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return orders, errors.Trasform(err)
	}

	return orders, nil
}

func (o *Order) Create(ctx context.Context, order entity.Order) error {
	query := `INSERT INTO orders (number, user_id, status) VALUES($1, $2, $3)`
	_, err := o.pool.Exec(ctx, query, order.Number, order.UserID, entity.OrderStatusNew)

	if err != nil {
		return errors.Trasform(err)
	}

	return nil
}
