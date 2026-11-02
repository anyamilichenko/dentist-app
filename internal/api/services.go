package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"dentist-app/internal/store"
)

var serviceErrs = errText{
	notFound:  "услуга не найдена",
	inUse:     "услуга используется в записях — её нельзя удалить",
	duplicate: "услуга с таким названием уже есть",
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.store.ListServices(r.Context())
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (h *Handler) getService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	sv, err := h.store.GetService(r.Context(), id)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	writeJSON(w, http.StatusOK, sv)
}

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {
	var in store.ServiceInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	if err := normalizeService(&in); err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	sv, err := h.store.CreateService(r.Context(), in)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	writeJSON(w, http.StatusCreated, sv)
}

func (h *Handler) updateService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	var in store.ServiceInput
	if err := decodeJSON(w, r, &in); err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	if err := normalizeService(&in); err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	sv, err := h.store.UpdateService(r.Context(), id, in)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	writeJSON(w, http.StatusOK, sv)
}

func (h *Handler) deleteService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	if err := h.store.DeleteService(r.Context(), id); err != nil {
		h.fail(w, r, err, serviceErrs)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func normalizeService(in *store.ServiceInput) error {
	in.Name = strings.TrimSpace(in.Name)
	switch {
	case in.Name == "":
		return validationError("укажите название услуги")
	case utf8.RuneCountInString(in.Name) > 200:
		return validationError("название услуги: не больше 200 символов")
	case in.DurationMinutes < 5 || in.DurationMinutes > 720:
		return validationError("длительность услуги: от 5 до 720 минут")
	case in.Price < 0:
		return validationError("цена не может быть отрицательной")
	}
	return nil
}
