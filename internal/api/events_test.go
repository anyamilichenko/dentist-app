package api

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEventsStream(t *testing.T) {
	broker := NewBroker()
	h := &Handler{broker: broker, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	// Через withLogging, как в настоящем роутере: проверяем, что Flush проходит сквозь обёртку.
	srv := httptest.NewServer(withLogging(h.log, http.HandlerFunc(h.events)))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q", ct)
	}
	lines := bufio.NewReader(resp.Body)
	expectLine(t, lines, "retry: 3000")

	broker.Publish()
	expectLine(t, lines, "data: changed")

	// После Close поток должен завершиться, иначе остановка сервера зависнет.
	broker.Close()
	if _, err := io.ReadAll(lines); err != nil {
		t.Fatalf("stream did not end cleanly: %v", err)
	}
}

func expectLine(t *testing.T, r *bufio.Reader, want string) {
	t.Helper()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("waiting for %q: %v", want, err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line != want {
			t.Fatalf("got %q, want %q", line, want)
		}
		return
	}
}

func TestNotifyChanges(t *testing.T) {
	broker := NewBroker()
	ch, unsubscribe := broker.subscribe()
	defer unsubscribe()

	status := http.StatusCreated
	handler := notifyChanges(broker, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	call := func(method string) bool {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, "/api/patients", nil))
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}

	if !call(http.MethodPost) {
		t.Error("successful POST should publish")
	}
	if call(http.MethodGet) {
		t.Error("GET should not publish")
	}
	status = http.StatusConflict
	if call(http.MethodPut) {
		t.Error("failed PUT should not publish")
	}
}
