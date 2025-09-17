package errors

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgErrCodeDuplicateKey = "23505"
	PgErrCodeNotFound     = "20000"
)

var (
	ErrDuplicateKey  = errors.New("duplicate key")
	ErrNotFound      = errors.New("not found")
	ErrNoRowsUpdated = errors.New("no rows updated")
)

func Trasform(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgErrCodeDuplicateKey:
			return ErrDuplicateKey
		case PgErrCodeNotFound:
			return ErrNotFound
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
