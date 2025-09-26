package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/pg"
)

type Processing struct {
	pool    *pgxpool.Pool
	delay   uint64
	retries uint64
}

func NewProcessing(db *pg.DB, delaySecond uint64, retryCount uint64) *Processing {
	return &Processing{
		pool:    db.Pool(),
		delay:   delaySecond,
		retries: retryCount,
	}
}

func (p *Processing) OrderNubmersForProcess(ctx context.Context, statuses []string) ([]string, error) {
	var numbers = []string{}
	pgStatuses := strings.Join(statuses, "', '")
	query := `UPDATE orders SET updated_at = NOW()
				WHERE status IN('%s') AND updated_at < NOW() - (attempts+1)*INTERVAL '%d seconds'
				RETURNING number`
	query = fmt.Sprintf(query, pgStatuses, p.delay)

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return numbers, fmt.Errorf("failed to select orders numbers for process: %w", err)
	}

	numbers, err = pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return numbers, fmt.Errorf("failed to parse selected orders: %w", err)
	}

	return numbers, nil
}

func (p *Processing) ProcessOrder(ctx context.Context, order entity.Order) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userID, err := updateOrderForProcess(ctx, tx, order)
	if err != nil {
		return err
	}

	if order.Status == entity.OrderStatusProcessed {
		err = increaseBalance(ctx, tx, order.Accrual, userID)
		if err != nil {
			return nil
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func updateOrderForProcess(ctx context.Context, tx pgx.Tx, o entity.Order) (userID uint64, err error) {
	query := `UPDATE orders SET status = $1, accrual = $2, updated_at = NOW(), attempts = 0
				WHERE number = $3 RETURNING user_id`

	rows, err := tx.Query(ctx, query, o.Status, o.Accrual, o.Number)
	if err != nil {
		return userID, fmt.Errorf("failed to update order#%s : %w", o.Number, err)
	}

	userID, err = pgx.CollectOneRow(rows, pgx.RowTo[uint64])
	if err != nil {
		return userID, fmt.Errorf("failed to parse user_id from order#%s : %w", o.Number, err)
	}

	return userID, err
}

func increaseBalance(ctx context.Context, tx pgx.Tx, accrual float64, userID uint64) error {
	_, err := tx.Exec(ctx, `LOCK TABLE balance IN ROW EXCLUSIVE MODE`)
	if err != nil {
		return fmt.Errorf("failed to lock balance table: %w", err)
	}

	query := `UPDATE balance SET balance = balance + $1 WHERE user_id = $2`
	tag, err := tx.Exec(ctx, query, accrual, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("failed to update balance: %w", errors.ErrNoRowsUpdated)
	}

	return nil
}

func (p *Processing) MarkOrderForRetryOrInvalid(
	ctx context.Context,
	number string,
	toStatus string,
	invalidStatus string,
) error {
	query := `UPDATE orders 
				SET 
					status = CASE WHEN attempts < $1 THEN $2 ELSE $3 END,
					attempts = attempts +1,
					updated_at = NOW()
				WHERE number = $4`

	tag, err := p.pool.Exec(ctx, query, p.retries, toStatus, invalidStatus, number)
	if err != nil {
		return fmt.Errorf("failed to update order#%s : %w", number, err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("failed to update order#%s : %w", number, errors.ErrNoRowsUpdated)
	}

	return nil
}
