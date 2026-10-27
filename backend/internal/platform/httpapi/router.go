package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"
)

// NewRouter returns the API handler. Unmatched routes and wrong methods
// respond with a JSON body of code and message.
//
// requestLog(jsonErrors(mux)) wraps the mux in two layers. A request hits
// requestLog first, then jsonErrors, then the route. The response comes back
// through those layers in reverse.
func NewRouter(log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	return requestLog(log, jsonErrors(mux))
}

// jsonErrors lets the inner handler write into a recorder, then copies the
// result onto the real response. A 404 or 405 from the mux is plain text;
// this replaces that body with JSON. Every other status is copied unchanged.
func jsonErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{header: make(http.Header)}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		dst := w.Header()
		for key, values := range rec.header {
			dst[key] = values
		}

		switch rec.status {
		case http.StatusNotFound:
			writeError(w, http.StatusNotFound, "not_found", "Not found")
		case http.StatusMethodNotAllowed:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		default:
			w.WriteHeader(rec.status)
			_, _ = w.Write(rec.body.Bytes())
		}
	})
}

// responseRecorder is a stand-in for the real response. It stores headers,
// status, and body in memory so jsonErrors can decide what to send.
type responseRecorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

// Header returns the header map the inner handler should fill.
// The name before Header is the receiver: this function belongs to *responseRecorder.
func (r *responseRecorder) Header() http.Header {
	return r.header
}

// WriteHeader records the status and does not send it yet.
// The first call wins. Later calls are ignored, matching net/http.
func (r *responseRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

// Write appends bytes to the stored body. Writing implies status 200 when
// the handler never called WriteHeader. (int, error) is the required return
// pair of http.ResponseWriter.
func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(p)
}

// requestLog times the inner handler and logs method, path, status, and duration.
// It returns a new handler; it does not replace next.
func requestLog(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start).String(),
		)
	})
}

// statusWriter wraps the real response so requestLog can read the status code.
// Embedding http.ResponseWriter means Header and other methods are reused;
// WriteHeader and Write below replace those two.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

// WriteHeader sends the status to the client and remembers it for the log.
// wrote blocks a second call, because net/http panics if WriteHeader runs twice.
func (w *statusWriter) WriteHeader(status int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write sends the body to the client. If the handler wrote bytes without a
// status, this records 200 first, which is what net/http does itself.
func (w *statusWriter) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}
