package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Appointment — запись на приём. Пациент и услуга приходят вместе с записью,
// чтобы календарю не нужно было делать отдельные запросы.
type Appointment struct {
	ID        int64     `json:"id"`
	PatientID int64     `json:"patient_id"`
	ServiceID *int64    `json:"service_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Patient   Patient   `json:"patient"`
	Service   *Service  `json:"service"`
}

type AppointmentInput struct {
	PatientID int64     `json:"patient_id"`
	ServiceID *int64    `json:"service_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Note      string    `json:"note"`
}

// AppointmentFilter: интервал [From, To) и/или пациент. Нулевые значения не фильтруют.
type AppointmentFilter struct {
	From      time.Time
	To        time.Time
	PatientID int64
}

// appointmentSelect — общий SELECT с пациентом и услугой; %s — таблица или CTE с записями.
const appointmentSelect = `
	SELECT a.id, a.patient_id, a.service_id, a.starts_at, a.ends_at, a.note, a.created_at, a.updated_at,
	       p.id, p.last_name, p.first_name, p.middle_name, p.phone, p.created_at, p.updated_at,
	       s.id, s.name, s.duration_minutes, s.price, s.created_at, s.updated_at
	FROM %s a
	JOIN patients p ON p.id = a.patient_id
	LEFT JOIN services s ON s.id = a.service_id`

func scanAppointment(row pgx.Row) (Appointment, error) {
	var (
		a  Appointment
		p  = &a.Patient
		sv struct {
			ID              *int64
			Name            *string
			DurationMinutes *int
			Price           *int
			CreatedAt       *time.Time
			UpdatedAt       *time.Time
		}
	)
	err := row.Scan(
		&a.ID, &a.PatientID, &a.ServiceID, &a.StartsAt, &a.EndsAt, &a.Note, &a.CreatedAt, &a.UpdatedAt,
		&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Phone, &p.CreatedAt, &p.UpdatedAt,
		&sv.ID, &sv.Name, &sv.DurationMinutes, &sv.Price, &sv.CreatedAt, &sv.UpdatedAt,
	)
	if err != nil {
		return Appointment{}, err
	}
	if sv.ID != nil {
		a.Service = &Service{
			ID:              *sv.ID,
			Name:            *sv.Name,
			DurationMinutes: *sv.DurationMinutes,
			Price:           *sv.Price,
			CreatedAt:       *sv.CreatedAt,
			UpdatedAt:       *sv.UpdatedAt,
		}
	}
	return a, nil
}

func (s *Store) ListAppointments(ctx context.Context, f AppointmentFilter) ([]Appointment, error) {
	var (
		where []string
		args  []any
	)
	if !f.From.IsZero() && !f.To.IsZero() {
		args = append(args, f.From, f.To)
		where = append(where, fmt.Sprintf("tstzrange(a.starts_at, a.ends_at) && tstzrange($%d, $%d)", len(args)-1, len(args)))
	}
	if f.PatientID != 0 {
		args = append(args, f.PatientID)
		where = append(where, fmt.Sprintf("a.patient_id = $%d", len(args)))
	}

	sql := fmt.Sprintf(appointmentSelect, "appointments")
	if len(where) > 0 {
		sql += ` WHERE ` + strings.Join(where, " AND ")
	}
	sql += ` ORDER BY a.starts_at, a.id`

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Appointment, error) { return scanAppointment(r) })
}

func (s *Store) GetAppointment(ctx context.Context, id int64) (Appointment, error) {
	a, err := scanAppointment(s.db.QueryRow(ctx, fmt.Sprintf(appointmentSelect, "appointments")+` WHERE a.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Appointment{}, ErrNotFound
	}
	return a, err
}

func (s *Store) CreateAppointment(ctx context.Context, in AppointmentInput) (Appointment, error) {
	a, err := scanAppointment(s.db.QueryRow(ctx, `
		WITH changed AS (
			INSERT INTO appointments (patient_id, service_id, starts_at, ends_at, note)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING *
		)`+fmt.Sprintf(appointmentSelect, "changed"),
		in.PatientID, in.ServiceID, in.StartsAt, in.EndsAt, in.Note))
	return a, appointmentWriteError(err)
}

func (s *Store) UpdateAppointment(ctx context.Context, id int64, in AppointmentInput) (Appointment, error) {
	a, err := scanAppointment(s.db.QueryRow(ctx, `
		WITH changed AS (
			UPDATE appointments
			SET patient_id = $1, service_id = $2, starts_at = $3, ends_at = $4, note = $5, updated_at = now()
			WHERE id = $6
			RETURNING *
		)`+fmt.Sprintf(appointmentSelect, "changed"),
		in.PatientID, in.ServiceID, in.StartsAt, in.EndsAt, in.Note, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Appointment{}, ErrNotFound
	}
	return a, appointmentWriteError(err)
}

func (s *Store) DeleteAppointment(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM appointments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func appointmentWriteError(err error) error {
	switch pgCode(err) {
	case codeExclusionViolation:
		return ErrOverlap
	case codeForeignKeyViolation:
		return ErrBadRef
	}
	return err
}
