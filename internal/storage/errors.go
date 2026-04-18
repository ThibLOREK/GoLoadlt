package storage

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func mapPGError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPipelineNotFound
	}
	return err
}
