package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"dentist-app/internal/store"
)

var patientErrs = errText{
	notFound: "пациент не найден",
	inUse:    "у пациента есть записи — сначала удалите их",
}

// GET /api/patients?q=иванов&limit=50&offset=0
func (h *Handler) listPatients(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := intParam(q.Get("limit"), 50)
	if err != nil || limit < 1 || limit > 200 {
		h.fail(w, r, validationError("limit должен быть от 1 до 200"), patientErrs)
		return
	}
	offset, err := intParam(q.Get("offset"), 0)
	if err != nil || offset < 0 {
		h.fail(w, r, validationError("offset должен быть неотрицательным"), patientErrs)
		return
	}

	patients, err := h.store.ListPatients(r.Context(), q.Get("q"), limit, offset)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	writeJSON(w, http.StatusOK, patients)
}

func (h *Handler) getPatient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	p, err := h.store.GetPatient(r.Context(), id)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) createPatient(w http.ResponseWriter, r *http.Request) {
	var in store.PatientInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	if err := normalizePatient(&in); err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	p, err := h.store.CreatePatient(r.Context(), in)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) updatePatient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	var in store.PatientInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	if err := normalizePatient(&in); err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	p, err := h.store.UpdatePatient(r.Context(), id, in)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deletePatient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	if err := h.store.DeletePatient(r.Context(), id); err != nil {
		h.fail(w, r, err, patientErrs)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var phoneRe = regexp.MustCompile(`^\+?[0-9()\- ]{5,32}$`)

func normalizePatient(in *store.PatientInput) error {
	in.LastName = strings.TrimSpace(in.LastName)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.MiddleName = trimOptional(in.MiddleName)
	in.Phone = trimOptional(in.Phone)

	switch {
	case in.LastName == "":
		return validationError("укажите фамилию")
	case in.FirstName == "":
		return validationError("укажите имя")
	case utf8.RuneCountInString(in.LastName) > 100,
		utf8.RuneCountInString(in.FirstName) > 100,
		in.MiddleName != nil && utf8.RuneCountInString(*in.MiddleName) > 100:
		return validationError("ФИО: не больше 100 символов в каждом поле")
	case in.Phone != nil && !phoneRe.MatchString(*in.Phone):
		return validationError("телефон: только цифры, пробелы и символы + - ( )")
	}
	return nil
}

// trimOptional обрезает пробелы; пустая строка превращается в nil (NULL в базе).
func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func intParam(s string, fallback int) (int, error) {
	if s == "" {
		return fallback, nil
	}
	return strconv.Atoi(s)
}
