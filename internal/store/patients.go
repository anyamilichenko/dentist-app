package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
)

type Patient struct {
	ID         int64     `json:"id"`
	LastName   string    `json:"last_name"`
	FirstName  string    `json:"first_name"`
	MiddleName *string   `json:"middle_name"`
	Phone      *string   `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PatientInput struct {
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	Phone      *string `json:"phone"`
}

const patientColumns = `id, last_name, first_name, middle_name, phone, created_at, updated_at`

func scanPatient(row pgx.Row) (Patient, error) {
	var p Patient
	err := row.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Phone, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// ListPatients ищет пациентов по ФИО и телефону. Каждое слово запроса должно
// встретиться хотя бы в одном поле: "иван 916" найдёт Иванова с номером +7 916 ...
func (s *Store) ListPatients(ctx context.Context, query string, limit, offset int) ([]Patient, error) {
	var (
		where []string
		args  []any
	)
	for _, word := range strings.Fields(query) {
		args = append(args, containsPattern(word))
		cond := fmt.Sprintf(
			"last_name ILIKE $%[1]d OR first_name ILIKE $%[1]d OR middle_name ILIKE $%[1]d OR phone ILIKE $%[1]d",
			len(args))
		if digits := onlyDigits(word); digits != "" {
			args = append(args, containsPattern(digits))
			cond += fmt.Sprintf(` OR regexp_replace(coalesce(phone, ''), '\D', '', 'g') LIKE $%d`, len(args))
		}
		where = append(where, "("+cond+")")
	}

	sql := `SELECT ` + patientColumns + ` FROM patients`
	if len(where) > 0 {
		sql += ` WHERE ` + strings.Join(where, " AND ")
	}
	args = append(args, limit, offset)
	sql += fmt.Sprintf(` ORDER BY last_name, first_name, middle_name NULLS FIRST, id LIMIT $%d OFFSET $%d`,
		len(args)-1, len(args))

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Patient, error) { return scanPatient(r) })
}

func (s *Store) GetPatient(ctx context.Context, id int64) (Patient, error) {
	p, err := scanPatient(s.db.QueryRow(ctx, `SELECT `+patientColumns+` FROM patients WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Patient{}, ErrNotFound
	}
	return p, err
}

func (s *Store) CreatePatient(ctx context.Context, in PatientInput) (Patient, error) {
	return scanPatient(s.db.QueryRow(ctx, `
		INSERT INTO patients (last_name, first_name, middle_name, phone)
		VALUES ($1, $2, $3, $4)
		RETURNING `+patientColumns,
		in.LastName, in.FirstName, in.MiddleName, in.Phone))
}

func (s *Store) UpdatePatient(ctx context.Context, id int64, in PatientInput) (Patient, error) {
	p, err := scanPatient(s.db.QueryRow(ctx, `
		UPDATE patients
		SET last_name = $1, first_name = $2, middle_name = $3, phone = $4, updated_at = now()
		WHERE id = $5
		RETURNING `+patientColumns,
		in.LastName, in.FirstName, in.MiddleName, in.Phone, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Patient{}, ErrNotFound
	}
	return p, err
}

func (s *Store) DeletePatient(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM patients WHERE id = $1`, id)
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

func onlyDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}
