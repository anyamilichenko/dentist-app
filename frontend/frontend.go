// Package frontend встраивает собранный интерфейс (npm run build → dist/app) в бинарник сервера,
// чтобы для работы хватало одного процесса и одного порта.
package frontend

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// dist/.keep лежит в репозитории, чтобы сервер собирался и без собранного интерфейса.
//
//go:embed all:dist
var dist embed.FS

// Handler отдаёт встроенный интерфейс или подсказку, если его ещё не собрали.
func Handler() http.Handler {
	files, err := fs.Sub(dist, "dist/app")
	if err != nil {
		return http.HandlerFunc(notBuilt)
	}
	return NewHandler(files)
}

// NewHandler отдаёт одностраничное приложение из files. Пути без файла (например, /patients) —
// это маршруты React Router, на них отдаётся index.html.
func NewHandler(files fs.FS) http.Handler {
	index, err := fs.ReadFile(files, "index.html")
	if err != nil {
		return http.HandlerFunc(notBuilt)
	}
	// index.html не кэшируем, чтобы после обновления сразу подхватывалась новая версия.
	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
	}

	fileServer := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "index.html" {
			serveIndex(w, r)
			return
		}
		if _, err := fs.Stat(files, name); err != nil {
			// Нет файла с расширением — это действительно 404 (например, старый ассет),
			// а не страница приложения: браузеру нельзя отдавать HTML вместо скрипта.
			if path.Ext(name) != "" {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, r)
			return
		}
		// У файлов в assets/ хеш в имени, их можно кэшировать навсегда.
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fileServer.ServeHTTP(w, r)
	})
}

func notBuilt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`<!doctype html><meta charset="utf-8"><title>Интерфейс не собран</title>
<body style="font-family: system-ui, sans-serif; max-width: 560px; margin: 48px auto; line-height: 1.5">
<h2>Интерфейс не собран</h2>
<p>API работает, но сборки фронтенда в бинарнике нет. Соберите её и перезапустите сервер:</p>
<pre style="background:#f4f6f8; padding:12px; border-radius:8px">cd frontend
npm run build
cd ..
go run ./cmd/server</pre>
<p>Или запустите dev-сервер интерфейса (<code>npm run dev</code>) и откройте <a href="http://localhost:5173">localhost:5173</a>.</p>
</body>`))
}
