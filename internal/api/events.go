package api

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// Broker рассылает открытым потокам /api/events сигнал «данные изменились».
// Что именно изменилось, не передаётся: клиент просто перезапрашивает то, что у него на экране.
type Broker struct {
	mu        sync.Mutex
	subs      map[chan struct{}]struct{}
	closed    chan struct{}
	closeOnce sync.Once
}

func NewBroker() *Broker {
	return &Broker{
		subs:   make(map[chan struct{}]struct{}),
		closed: make(chan struct{}),
	}
}

func (b *Broker) subscribe() (ch chan struct{}, unsubscribe func()) {
	ch = make(chan struct{}, 1)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.subs, ch)
		b.mu.Unlock()
	}
}

// Publish будит всех подписчиков. Сигналы не копятся: если подписчик ещё не забрал
// предыдущий, новый не нужен — одного «изменилось» достаточно.
func (b *Broker) Publish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Close завершает все открытые потоки; вызывается при остановке сервера,
// иначе Shutdown ждал бы бесконечные SSE-соединения.
func (b *Broker) Close() {
	b.closeOnce.Do(func() { close(b.closed) })
}

const heartbeatInterval = 25 * time.Second

// GET /api/events — поток Server-Sent Events: "data: changed" после любого изменения данных.
func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	// Чтобы nginx перед сервером (при деплое) не буферизовал поток.
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsubscribe := h.broker.subscribe()
	defer unsubscribe()

	// retry — через сколько миллисекунд браузер переподключится после обрыва.
	if !send(w, rc, "retry: 3000\n\n") {
		return
	}

	// Комментарий-пинг держит соединение живым через прокси и помогает заметить отвалившихся клиентов.
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		var msg string
		select {
		case <-r.Context().Done():
			return
		case <-h.broker.closed:
			return
		case <-ch:
			msg = "data: changed\n\n"
		case <-heartbeat.C:
			msg = ": ping\n\n"
		}
		if !send(w, rc, msg) {
			return
		}
	}
}

func send(w io.Writer, rc *http.ResponseController, msg string) bool {
	if _, err := io.WriteString(w, msg); err != nil {
		return false
	}
	return rc.Flush() == nil
}

// notifyChanges публикует событие после каждого успешного изменяющего запроса,
// чтобы новые обработчики не нужно было помнить подключать вручную.
func notifyChanges(b *Broker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status < 300 {
			b.Publish()
		}
	})
}
