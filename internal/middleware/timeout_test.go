package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutCompletesNormally(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	timeoutMiddleware := Timeout(100 * time.Millisecond)
	wrapped := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}

	if res.Body.String() != "ok" {
		t.Errorf("want 'ok', got %q", res.Body.String())
	}
}

func TestTimeoutExceeds(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	timeoutMiddleware := Timeout(10 * time.Millisecond)
	wrapped := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusGatewayTimeout {
		t.Errorf("want 504, got %d", res.Code)
	}

	expected := "Gateway Timeout\n"
	if res.Body.String() != expected {
		t.Errorf("want %q, got %q", expected, res.Body.String())
	}
}

func TestTimeoutConfigurable(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	short := Timeout(20 * time.Millisecond)
	wrappedShort := short(handler)

	req := httptest.NewRequest("GET", "/", nil)
	resShort := httptest.NewRecorder()
	wrappedShort.ServeHTTP(resShort, req)

	if resShort.Code != http.StatusGatewayTimeout {
		t.Errorf("short timeout: want 504, got %d", resShort.Code)
	}

	long := Timeout(200 * time.Millisecond)
	wrappedLong := long(handler)

	resLong := httptest.NewRecorder()
	wrappedLong.ServeHTTP(resLong, req)

	if resLong.Code != http.StatusOK {
		t.Errorf("long timeout: want 200, got %d", resLong.Code)
	}
}

func TestTimeoutRespectsExistingDeadline(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("expected deadline to exist")
		}
		if time.Until(deadline) > 30*time.Millisecond {
			t.Error("expected shorter deadline from parent")
		}
		w.WriteHeader(http.StatusOK)
	})

	parentCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	timeoutMiddleware := Timeout(100 * time.Millisecond)
	wrapped := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil).WithContext(parentCtx)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestTimeoutCancelsContext(t *testing.T) {
	cancelled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		cancelled = true
		w.WriteHeader(http.StatusOK)
	})

	timeoutMiddleware := Timeout(10 * time.Millisecond)
	wrapped := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	time.Sleep(20 * time.Millisecond)

	if !cancelled {
		t.Error("expected context to be cancelled on timeout")
	}

	if res.Code != http.StatusGatewayTimeout {
		t.Errorf("want 504, got %d", res.Code)
	}
}

func TestTimeoutZeroDuration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	timeoutMiddleware := Timeout(0)
	wrapped := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	wrapped.ServeHTTP(res, req)

	if res.Code != http.StatusGatewayTimeout {
		t.Errorf("want 504, got %d", res.Code)
	}
}
