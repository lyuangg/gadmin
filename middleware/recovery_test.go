package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// mockLogger 记录 ErrorContext 调用，便于断言日志内容
type mockLogger struct {
	mu         sync.Mutex
	errorCalls []errorCall
}

type errorCall struct {
	Ctx  context.Context
	Msg  string
	Args []any
}

func (m *mockLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (m *mockLogger) WarnContext(ctx context.Context, msg string, args ...any) {}
func (m *mockLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {}

func (m *mockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, errorCall{Ctx: ctx, Msg: msg, Args: args})
}

func (m *mockLogger) getErrorCalls() []errorCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]errorCall(nil), m.errorCalls...)
}

// argsToMap 将 slog 风格的 key, value, key, value 转为 map[string]any
func argsToMap(args []any) map[string]any {
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

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	a := app.NewTestAppWithServiceMocks(nil)
	r := gin.New()
	r.Use(RecoveryMiddleware(a))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (Responder.Error 写的是 200+JSON)", rec.Code)
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != errors.CodeInternalError {
		t.Errorf("code = %d, want %d", body.Code, errors.CodeInternalError)
	}
	if body.Msg != "服务器内部错误" {
		t.Errorf("msg = %q, want %q", body.Msg, "服务器内部错误")
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	a := app.NewTestAppWithServiceMocks(nil)
	r := gin.New()
	r.Use(RecoveryMiddleware(a))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestRecoveryMiddleware_LogsContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockLogger{}
	a := app.NewTestAppWithLogger(mock)
	r := gin.New()
	r.Use(RecoveryMiddleware(a))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic value")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	calls := mock.getErrorCalls()
	if len(calls) != 1 {
		t.Fatalf("ErrorContext 调用次数 = %d, want 1", len(calls))
	}
	call := calls[0]
	if call.Msg != "panic recovered" {
		t.Errorf("msg = %q, want %q", call.Msg, "panic recovered")
	}
	m := argsToMap(call.Args)
	if m["panic"] != "test panic value" {
		t.Errorf("args[panic] = %v, want %q", m["panic"], "test panic value")
	}
	if m["path"] != "/panic" {
		t.Errorf("args[path] = %v, want %q", m["path"], "/panic")
	}
	if m["method"] != "GET" {
		t.Errorf("args[method] = %v, want GET", m["method"])
	}
	if _, ok := m["client_ip"]; !ok {
		t.Error("args 应包含 client_ip")
	}
	stack, ok := m["stack"].(string)
	if !ok || stack == "" {
		t.Errorf("args[stack] 应为非空字符串, got %v", m["stack"])
	}

	// 响应仍为统一错误格式
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != errors.CodeInternalError || body.Msg != "服务器内部错误" {
		t.Errorf("body = code %d msg %q, want code %d msg 服务器内部错误", body.Code, body.Msg, errors.CodeInternalError)
	}
}
