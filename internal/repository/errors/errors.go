package errors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgErrCodeDuplicateKey = "23505"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
)

func Trasform(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgErrCodeDuplicateKey:
			return ErrDuplicateKey
		}
	}

	return err
}
