package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTraceIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID string
		wantOK bool
	}{
		{
			name:   "nil context",
			ctx:    nil,
			wantID: "",
			wantOK: false,
		},
		{
			name:   "context without trace_id",
			ctx:    context.Background(),
			wantID: "",
			wantOK: false,
		},
		{
			name:   "context with empty string",
			ctx:    context.WithValue(context.Background(), traceIDContextKey, ""),
			wantID: "",
			wantOK: false,
		},
		{
			name:   "context with valid trace_id",
			ctx:    context.WithValue(context.Background(), traceIDContextKey, "abc123"),
			wantID: "abc123",
			wantOK: true,
		},
		{
			name:   "context with wrong type",
			ctx:    context.WithValue(context.Background(), traceIDContextKey, 42),
			wantID: "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := TraceIDFromContext(tt.ctx)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("TraceIDFromContext() = (%q, %v), want (%q, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestTraceIDMiddleware_usesHeaderWhenPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(traceIDHeader, "client-trace-id-12345")
	c.Request = req

	h := TraceIDMiddleware()
	h(c)

	if got := w.Header().Get(traceIDHeader); got != "client-trace-id-12345" {
		t.Errorf("response X-Trace-Id = %q, want %q", got, "client-trace-id-12345")
	}
	if got, ok := c.Get(traceIDContextKey); !ok || got != "client-trace-id-12345" {
		t.Errorf("c.Get(trace_id) = %v, %v, want client-trace-id-12345, true", got, ok)
	}
	if got, ok := TraceIDFromContext(c.Request.Context()); !ok || got != "client-trace-id-12345" {
		t.Errorf("TraceIDFromContext(request) = %q, %v, want client-trace-id-12345, true", got, ok)
	}
}

func TestTraceIDMiddleware_generatesIDWhenHeaderMissing(t *testing.T) {
	hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	h := TraceIDMiddleware()
	h(c)

	got := w.Header().Get(traceIDHeader)
	if got == "" {
		t.Fatal("response X-Trace-Id is empty")
	}
	if !hex32.MatchString(got) {
		t.Errorf("X-Trace-Id %q is not 32 lowercase hex chars", got)
	}
	if id, ok := c.Get(traceIDContextKey); !ok {
		t.Errorf("c.Get(trace_id) missing")
	} else if id != got {
		t.Errorf("c.Get(trace_id) = %q, want %q", id, got)
	}
	if id, ok := TraceIDFromContext(c.Request.Context()); !ok || id != got {
		t.Errorf("TraceIDFromContext(request) = %q, %v, want %q, true", id, ok, got)
	}
}

func TestTraceIDMiddleware_callsNext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	nextCalled := false
	e.Use(TraceIDMiddleware())
	e.GET("/", func(c *gin.Context) {
		nextCalled = true
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("middleware did not call next handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
