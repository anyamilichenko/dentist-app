package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandler(t *testing.T) {
	files := fstest.MapFS{
		"index.html":       {Data: []byte("<html>app</html>")},
		"favicon.svg":      {Data: []byte("<svg/>")},
		"assets/index.js":  {Data: []byte("console.log(1)")},
		"assets/index.css": {Data: []byte("body{}")},
	}
	h := NewHandler(files)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{"/", http.StatusOK, "app", "no-cache"},
		{"/patients", http.StatusOK, "app", "no-cache"},   // маршрут React Router
		{"/services/", http.StatusOK, "app", "no-cache"},  // со слэшем на конце
		{"/index.html", http.StatusOK, "app", "no-cache"}, // без редиректа
		{"/favicon.svg", http.StatusOK, "<svg/>", ""},     // обычный файл
		{"/assets/index.js", http.StatusOK, "console", "immutable"},
		{"/assets/old-hash.js", http.StatusNotFound, "", ""}, // не HTML вместо скрипта
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != tt.wantStatus {
			t.Errorf("%s: status %d, want %d", tt.path, rec.Code, tt.wantStatus)
			continue
		}
		if !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("%s: body %q does not contain %q", tt.path, rec.Body.String(), tt.wantBody)
		}
		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, tt.wantCache) {
			t.Errorf("%s: Cache-Control %q does not contain %q", tt.path, cc, tt.wantCache)
		}
	}
}
