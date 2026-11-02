package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"dentist-app/internal/store"
)

type Handler struct {
	store  *store.Store
	broker *Broker
	log    *slog.Logger
}

func NewRouter(st *store.Store, broker *Broker, log *slog.Logger) http.Handler {
	h := &Handler{store: st, broker: broker, log: log}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/events", h.events)

	mux.HandleFunc("GET /api/patients", h.listPatients)
	mux.HandleFunc("POST /api/patients", h.createPatient)
	mux.HandleFunc("GET /api/patients/{id}", h.getPatient)
	mux.HandleFunc("PUT /api/patients/{id}", h.updatePatient)
	mux.HandleFunc("DELETE /api/patients/{id}", h.deletePatient)

	mux.HandleFunc("GET /api/services", h.listServices)
	mux.HandleFunc("POST /api/services", h.createService)
	mux.HandleFunc("GET /api/services/{id}", h.getService)
	mux.HandleFunc("PUT /api/services/{id}", h.updateService)
	mux.HandleFunc("DELETE /api/services/{id}", h.deleteService)

	mux.HandleFunc("GET /api/appointments", h.listAppointments)
	mux.HandleFunc("POST /api/appointments", h.createAppointment)
	mux.HandleFunc("GET /api/appointments/{id}", h.getAppointment)
	mux.HandleFunc("PUT /api/appointments/{id}", h.updateAppointment)
	mux.HandleFunc("DELETE /api/appointments/{id}", h.deleteAppointment)

	return withLogging(log, notifyChanges(broker, mux))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// validationError — ошибка во входных данных; текст показывается пользователю как есть.
type validationError string

func (e validationError) Error() string { return string(e) }

// errText — тексты ошибок хранилища для конкретной сущности.
type errText struct {
	notFound  string
	inUse     string
	duplicate string
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error, t errText) {
	var verr validationError
	switch {
	case errors.As(err, &verr):
		writeError(w, http.StatusBadRequest, verr.Error())
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, t.notFound)
	case errors.Is(err, store.ErrInUse):
		writeError(w, http.StatusConflict, t.inUse)
	case errors.Is(err, store.ErrDuplicate):
		writeError(w, http.StatusConflict, t.duplicate)
	case errors.Is(err, store.ErrOverlap):
		writeError(w, http.StatusConflict, "это время уже занято другой записью")
	case errors.Is(err, store.ErrBadRef):
		writeError(w, http.StatusUnprocessableEntity, "пациент или услуга не найдены")
	default:
		h.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return validationError("некорректный JSON: " + err.Error())
	}
	return nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, validationError("некорректный id")
	}
	return id, nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Unwrap нужен http.ResponseController, чтобы добраться до Flush исходного соединения (для SSE).
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func withLogging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if p := recover(); p != nil {
				log.Error("panic", "err", p, "stack", string(debug.Stack()))
				writeError(rec, http.StatusInternalServerError, "внутренняя ошибка сервера")
			}
			log.Info("request", "method", r.Method, "path", r.URL.Path,
				"status", rec.status, "duration", time.Since(start).Round(time.Microsecond))
		}()
		next.ServeHTTP(rec, r)
	})
}
