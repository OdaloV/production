package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// minSize is the minimum response size to compress (512 bytes)
const minSize = 512

// compressibleTypes are content types that should be compressed
var compressibleTypes = map[string]bool{
	"text/html":          true,
	"text/plain":         true,
	"text/css":           true,
	"text/javascript":    true,
	"application/json":   true,
	"application/xml":    true,
	"application/javascript": true,
	"image/svg+xml":      true,
	"font/woff2":         false, // already compressed
	"image/jpeg":         false, // already compressed
	"image/png":          false, // already compressed
	"video/mp4":          false, // already compressed
}

// compress middleware compresses responses with gzip when supported
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// wrap response writer with gzip
		gzw := newGzipResponseWriter(w)
		defer gzw.close()

		// add compression headers
		gzw.Header().Set("Content-Encoding", "gzip")
		gzw.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
	// track if headers have been written
	wroteHeader bool
	// store status code
	status int
	// buffer small responses to check size
	buffer []byte
}

// newGzipResponseWriter creates a new gzip response writer
func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
		gzw:            gzip.NewWriter(w),
		status:         http.StatusOK,
		buffer:         make([]byte, 0, minSize),
	}
}

// WriteHeader captures status code
func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

// Write compresses or buffers response data
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(w.status)
	}

	// skip compression for small responses or certain status codes
	if w.shouldSkipCompression() {
		return w.ResponseWriter.Write(data)
	}

	// buffer until we reach minSize or know content type
	w.buffer = append(w.buffer, data...)

	// check if we should compress
	if len(w.buffer) >= minSize || w.isFlushable() {
		// set content encoding before writing headers
		w.Header().Set("Content-Encoding", "gzip")

		// write original headers
		for k, v := range w.ResponseWriter.Header() {
			w.ResponseWriter.Header()[k] = v
		}
		w.ResponseWriter.WriteHeader(w.status)

		// compress and write buffered data
		_, err := w.gzw.Write(w.buffer)
		if err != nil {
			return 0, err
		}
		w.buffer = nil

		// future writes go directly to gzip
		return len(data), nil
	}

	return len(data), nil
}

// shouldSkipCompression checks if compression should be skipped
func (w *gzipResponseWriter) shouldSkipCompression() bool {
	// don't compress 1xx, 204, 304 responses
	if w.status < 200 || w.status == 204 || w.status == 304 {
		return true
	}

	// check content type
	contentType := w.Header().Get("Content-Type")
	for ct, compress := range compressibleTypes {
		if strings.Contains(contentType, ct) && !compress {
			return true
		}
	}

	return false
}

// isFlushable checks if we have enough info to decide
func (w *gzipResponseWriter) isFlushable() bool {
	contentType := w.Header().Get("Content-Type")
	if contentType != "" {
		for ct, compress := range compressibleTypes {
			if strings.Contains(contentType, ct) {
				return compress
			}
		}
	}
	return len(w.buffer) >= minSize
}

// close flushes gzip writer
func (w *gzipResponseWriter) close() {
	if w.gzw != nil {
		w.gzw.Flush()
		w.gzw.Close()
	}
}

// Flush implements http.Flusher
func (w *gzipResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// compressWriter ensures gzipResponseWriter implements http.ResponseWriter
var _ http.ResponseWriter = (*gzipResponseWriter)(nil)
