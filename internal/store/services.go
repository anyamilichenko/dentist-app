package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           int       `json:"price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ServiceInput struct {
	Name            string `json:"name"`
	DurationMinutes int    `json:"duration_minutes"`
	Price           int    `json:"price"`
}

const serviceColumns = `id, name, duration_minutes, price, created_at, updated_at`

func scanService(row pgx.Row) (Service, error) {
	var sv Service
	err := row.Scan(&sv.ID, &sv.Name, &sv.DurationMinutes, &sv.Price, &sv.CreatedAt, &sv.UpdatedAt)
	return sv, err
}

func (s *Store) ListServices(ctx context.Context) ([]Service, error) {
	rows, err := s.db.Query(ctx, `SELECT `+serviceColumns+` FROM services ORDER BY name, id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Service, error) { return scanService(r) })
}

func (s *Store) GetService(ctx context.Context, id int64) (Service, error) {
	sv, err := scanService(s.db.QueryRow(ctx, `SELECT `+serviceColumns+` FROM services WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Service{}, ErrNotFound
	}
	return sv, err
}

func (s *Store) CreateService(ctx context.Context, in ServiceInput) (Service, error) {
	sv, err := scanService(s.db.QueryRow(ctx, `
		INSERT INTO services (name, duration_minutes, price)
		VALUES ($1, $2, $3)
		RETURNING `+serviceColumns,
		in.Name, in.DurationMinutes, in.Price))
	if pgCode(err) == codeUniqueViolation {
		return Service{}, ErrDuplicate
	}
	return sv, err
}

func (s *Store) UpdateService(ctx context.Context, id int64, in ServiceInput) (Service, error) {
	sv, err := scanService(s.db.QueryRow(ctx, `
		UPDATE services
		SET name = $1, duration_minutes = $2, price = $3, updated_at = now()
		WHERE id = $4
		RETURNING `+serviceColumns,
		in.Name, in.DurationMinutes, in.Price, id))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return Service{}, ErrNotFound
	case pgCode(err) == codeUniqueViolation:
		return Service{}, ErrDuplicate
	}
	return sv, err
}

func (s *Store) DeleteService(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	if pgCode(err) == codeForeignKeyViolation {
		return ErrInUse
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
