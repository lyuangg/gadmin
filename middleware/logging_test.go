package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lyuangg/gadmin/app"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// loggingMock 记录 InfoContext 调用，便于断言日志内容
type loggingMock struct {
	mu        sync.Mutex
	infoCalls []infoCall
}

type infoCall struct {
	Msg  string
	Args []any
}

func (m *loggingMock) InfoContext(_ context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalls = append(m.infoCalls, infoCall{Msg: msg, Args: args})
}
func (m *loggingMock) ErrorContext(context.Context, string, ...any) {}
func (m *loggingMock) WarnContext(context.Context, string, ...any)  {}
func (m *loggingMock) Log(context.Context, slog.Level, string, ...any) {}

func (m *loggingMock) getInfoCalls() []infoCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]infoCall(nil), m.infoCalls...)
}

func loggingArgsToMap(args []any) map[string]any {
	out := make(map[string]any)
	for i := 0; i+1 < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		out[k] = args[i+1]
	}
	return out
}

// 请求 /static/ 不记录日志（提前 return）
func TestLoggingMiddleware_SkipsStatic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &loggingMock{}
	a := app.NewTestAppWithLogger(mock)
	r := gin.New()
	r.Use(LoggingMiddleware(a))
	r.GET("/static/css/foo.css", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/static/css/foo.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	calls := mock.getInfoCalls()
	if len(calls) != 0 {
		t.Errorf("请求 /static/ 不应打日志, got %d calls", len(calls))
	}
}

// 非 API 请求记录 [PAGE] 日志，含 method、path、status、duration_ms、client_ip
func TestLoggingMiddleware_LogsPageRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &loggingMock{}
	a := app.NewTestAppWithLogger(mock)
	r := gin.New()
	r.Use(LoggingMiddleware(a))
	r.GET("/login", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	calls := mock.getInfoCalls()
	if len(calls) != 1 {
		t.Fatalf("InfoContext 调用次数 = %d, want 1", len(calls))
	}
	call := calls[0]
	if !strings.HasPrefix(call.Msg, "[PAGE] ") {
		t.Errorf("msg = %q, want prefix [PAGE] ", call.Msg)
	}
	if !strings.Contains(call.Msg, "GET") || !strings.Contains(call.Msg, "/login") || !strings.Contains(call.Msg, "200") {
		t.Errorf("msg = %q", call.Msg)
	}
	m := loggingArgsToMap(call.Args)
	if m["method"] != "GET" || m["path"] != "/login" {
		t.Errorf("attrs = %v", m)
	}
	if _, ok := m["duration_ms"]; !ok {
		t.Error("attrs 应包含 duration_ms")
	}
	if _, ok := m["req_body"]; ok {
		t.Error("PAGE 请求不应包含 req_body")
	}
}

// API 请求记录 [API] 日志，且含 req_body、resp_body
func TestLoggingMiddleware_LogsAPIRequestWithBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &loggingMock{}
	a := app.NewTestAppWithLogger(mock)
	r := gin.New()
	r.Use(LoggingMiddleware(a))
	r.POST("/api/foo", func(c *gin.Context) {
		c.String(200, `{"done":true}`)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/foo", strings.NewReader(`{"input":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	calls := mock.getInfoCalls()
	if len(calls) != 1 {
		t.Fatalf("InfoContext 调用次数 = %d, want 1", len(calls))
	}
	call := calls[0]
	if !strings.HasPrefix(call.Msg, "[API] ") {
		t.Errorf("msg = %q, want prefix [API] ", call.Msg)
	}
	m := loggingArgsToMap(call.Args)
	if m["req_body"] != `{"input":1}` {
		t.Errorf("req_body = %v", m["req_body"])
	}
	if m["resp_body"] != `{"done":true}` {
		t.Errorf("resp_body = %v", m["resp_body"])
	}
}
