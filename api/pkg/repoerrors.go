package pkg

import (
	"errors"
	"fmt"

	"atom-maintenance/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func MapDB(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrNotFound

	case isUniqueViolation(err):
		return domain.ErrConflict

	case isForeignKeyViolation(err):
		return domain.ErrBadRequest

	case isCheckViolation(err):
		return domain.ErrBadRequest

	default:
		return fmt.Errorf("db error: %w", err)
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}
