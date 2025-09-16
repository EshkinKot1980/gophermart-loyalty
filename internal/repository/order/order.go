package order

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
)

type Order struct {
	pool *pgxpool.Pool
}

func New(p *pgxpool.Pool) *Order {
	return &Order{pool: p}
}

func (o *Order) GetByNumber(ctx context.Context, number string) (entity.Order, error) {
	var order entity.Order
	query := `SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE number = $1`
	row := o.pool.QueryRow(ctx, query, number)

	err := row.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.Uploaded, &order.Updated)
	if err != nil {
		return entity.Order{}, errors.Trasform(err)
	}

	return order, nil
}

func (o *Order) Create(ctx context.Context, order entity.Order) error {
	query := `INSERT INTO orders (number, user_id, status) VALUES($1, $2, $3)`
	_, err := o.pool.Exec(ctx, query, order.Number, order.UserID, entity.OrderStatusNew)

	if err != nil {
		return errors.Trasform(err)
	}

	return nil
}
