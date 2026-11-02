package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"dentist-app/internal/store"
)

var appointmentErrs = errText{
	notFound: "запись не найдена",
}

// GET /api/appointments?from=2026-11-02T00:00:00Z&to=2026-11-09T00:00:00Z
// GET /api/appointments?patient_id=1
func (h *Handler) listAppointments(w http.ResponseWriter, r *http.Request) {
	f, err := parseAppointmentFilter(r)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	list, err := h.store.ListAppointments(r.Context(), f)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	a, err := h.store.GetAppointment(r.Context(), id)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) createAppointment(w http.ResponseWriter, r *http.Request) {
	var in store.AppointmentInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	if err := normalizeAppointment(&in); err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	a, err := h.store.CreateAppointment(r.Context(), in)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) updateAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	var in store.AppointmentInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	if err := normalizeAppointment(&in); err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	a, err := h.store.UpdateAppointment(r.Context(), id, in)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) deleteAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	if err := h.store.DeleteAppointment(r.Context(), id); err != nil {
		h.fail(w, r, err, appointmentErrs)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseAppointmentFilter(r *http.Request) (store.AppointmentFilter, error) {
	var f store.AppointmentFilter
	q := r.URL.Query()

	var err error
	if f.From, err = timeParam(q.Get("from")); err != nil {
		return f, validationError("from: ожидается дата в формате RFC 3339, например 2026-11-02T09:00:00+03:00")
	}
	if f.To, err = timeParam(q.Get("to")); err != nil {
		return f, validationError("to: ожидается дата в формате RFC 3339, например 2026-11-09T00:00:00+03:00")
	}
	if v := q.Get("patient_id"); v != "" {
		if f.PatientID, err = strconv.ParseInt(v, 10, 64); err != nil || f.PatientID <= 0 {
			return f, validationError("некорректный patient_id")
		}
	}

	switch {
	case f.From.IsZero() != f.To.IsZero():
		return f, validationError("укажите оба параметра: from и to")
	case f.From.IsZero() && f.PatientID == 0:
		return f, validationError("укажите период (from и to) или patient_id")
	case !f.From.IsZero() && !f.To.After(f.From):
		return f, validationError("to должен быть позже from")
	case f.To.Sub(f.From) > 92*24*time.Hour:
		return f, validationError("период не может быть длиннее 3 месяцев")
	}
	return f, nil
}

func timeParam(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

func normalizeAppointment(in *store.AppointmentInput) error {
	in.Note = strings.TrimSpace(in.Note)
	switch {
	case in.PatientID <= 0:
		return validationError("выберите пациента")
	case in.ServiceID != nil && *in.ServiceID <= 0:
		return validationError("некорректная услуга")
	case in.StartsAt.IsZero():
		return validationError("укажите время начала")
	case in.EndsAt.IsZero():
		return validationError("укажите время окончания")
	case !in.EndsAt.After(in.StartsAt):
		return validationError("время окончания должно быть позже начала")
	case in.EndsAt.Sub(in.StartsAt) > 12*time.Hour:
		return validationError("запись не может быть длиннее 12 часов")
	case utf8.RuneCountInString(in.Note) > 2000:
		return validationError("комментарий: не больше 2000 символов")
	}
	return nil
}
