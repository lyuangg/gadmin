package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/internal/testutil"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

// 非 /admin/api/ 路径不记录，直接 Next
func TestOperationLogMiddleware_SkipsNonAdminAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestDB(t)
	a := app.NewTestApp(db)
	r := gin.New()
	r.Use(OperationLogMiddleware(a))
	r.GET("/api/foo", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	// 不记录时不会写 operation_logs，可查表确认条数为 0
	var count int64
	db.Model(&models.OperationLog{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 logs for non-admin/api path, got %d", count)
	}
}

// GET 方法不记录（只记录 PUT/DELETE/POST）
func TestOperationLogMiddleware_SkipsGetMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestDB(t)
	a := app.NewTestApp(db)
	r := gin.New()
	r.Use(OperationLogMiddleware(a))
	r.GET("/admin/api/users", func(c *gin.Context) { c.String(200, "[]") })

	req := httptest.NewRequest(http.MethodGet, "/admin/api/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var count int64
	db.Model(&models.OperationLog{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 logs for GET, got %d", count)
	}
}

// POST /admin/api/ 且 context 有 claims 时异步写入一条操作日志
func TestOperationLogMiddleware_RecordsLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTestDB(t)
	a := app.NewTestApp(db)
	r := gin.New()
	// 模拟 JWT 之后设置 claims，供中间件读取
	r.Use(func(c *gin.Context) {
		c.Set("claims", &utils.Claims{UserID: 1, Username: "oploguser"})
		c.Next()
	})
	r.Use(OperationLogMiddleware(a))
	r.POST("/admin/api/roles", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": 1})
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/api/roles", strings.NewReader(`{"name":"role1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	// 中间件在 goroutine 里写库，稍等再查
	time.Sleep(100 * time.Millisecond)

	var logs []models.OperationLog
	if err := db.Order("id DESC").Limit(1).Find(&logs).Error; err != nil {
		t.Fatalf("find logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	got := logs[0]
	if got.Method != "POST" || got.Path != "/admin/api/roles" {
		t.Errorf("method=%s path=%s", got.Method, got.Path)
	}
	if got.UserID != 1 || got.Username != "oploguser" {
		t.Errorf("user_id=%d username=%s", got.UserID, got.Username)
	}
	if got.StatusCode != 200 {
		t.Errorf("status_code=%d", got.StatusCode)
	}
	if !strings.Contains(got.Request, "role1") {
		t.Errorf("request should contain role1, got %s", got.Request)
	}
	if !strings.Contains(got.Response, "id") {
		t.Errorf("response should contain id, got %s", got.Response)
	}
}
