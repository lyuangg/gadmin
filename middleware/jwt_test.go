package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-jwt-secret-for-auth-middleware"

func initTestJWT(t *testing.T) {
	t.Helper()
	utils.InitJWT(&config.Config{JWTSecret: testJWTSecret})
}

// 无 token 时 API 请求应返回 401 统一错误
func TestAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	a := app.NewTestAppWithServiceMocks(nil)
	r := gin.New()
	r.Use(AuthMiddleware(a))
	r.GET("/api/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (Responder 写 JSON 时用 200)", rec.Code)
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != errors.CodeUnauthorized {
		t.Errorf("code = %d, want %d", body.Code, errors.CodeUnauthorized)
	}
}

// 无效 token 时 API 请求应返回 401
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestJWT(t)
	a := app.NewTestAppWithServiceMocks(nil)
	r := gin.New()
	r.Use(AuthMiddleware(a))
	r.GET("/api/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != errors.CodeUnauthorized {
		t.Errorf("code = %d, want %d", body.Code, errors.CodeUnauthorized)
	}
}

// 有效 token 但 GetUserForAuth 返回错误（用户不存在）→ 401
func TestAuthMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestJWT(t)
	token, err := utils.GenerateToken(1, "u", "n", 0, false, []uint{1}, 0)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	userMock := &services.FakeUserService{
		GetUserForAuthUser: nil,
		GetUserForAuthErr:  errors.UnauthorizedMsg("用户不存在"),
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	r := gin.New()
	r.Use(AuthMiddleware(a))
	r.GET("/api/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != errors.CodeUnauthorized {
		t.Errorf("code = %d, want %d", body.Code, errors.CodeUnauthorized)
	}
}

// 有效 token 但 token_version 小于用户当前版本 → 401
func TestAuthMiddleware_TokenVersionMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestJWT(t)
	// token 里 token_version=0，用户当前 token_version=1
	token, err := utils.GenerateToken(1, "u", "n", 0, false, []uint{1}, 0)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	userMock := &services.FakeUserService{
		GetUserForAuthUser: &models.User{ID: 1, Username: "u", Nickname: "n", Type: 0, Status: 1, TokenVersion: 1},
		GetUserForAuthErr:  nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	r := gin.New()
	r.Use(AuthMiddleware(a))
	r.GET("/api/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body struct {
		Code int `json:"code"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != errors.CodeUnauthorized {
		t.Errorf("code = %d, want %d (Token已失效)", body.Code, errors.CodeUnauthorized)
	}
}

// 有效 token 但用户已禁用 → 403
func TestAuthMiddleware_UserDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestJWT(t)
	token, err := utils.GenerateToken(1, "u", "n", 0, false, []uint{1}, 0)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	userMock := &services.FakeUserService{
		GetUserForAuthUser: &models.User{ID: 1, Username: "u", Nickname: "n", Type: 0, Status: 0, TokenVersion: 0},
		GetUserForAuthErr:  nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	r := gin.New()
	r.Use(AuthMiddleware(a))
	r.GET("/api/protected", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body struct {
		Code int `json:"code"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != errors.CodeForbidden {
		t.Errorf("code = %d, want %d (用户已被禁用)", body.Code, errors.CodeForbidden)
	}
}

// 有效 token + 用户存在且启用 → 通过，user 与 claims 写入上下文
func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestJWT(t)
	token, err := utils.GenerateToken(1, "testuser", "测试", 0, false, []uint{1, 2}, 0)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	userMock := &services.FakeUserService{
		GetUserForAuthUser: &models.User{ID: 1, Username: "testuser", Nickname: "测试", Type: 0, Status: 1, TokenVersion: 0},
		GetUserForAuthErr:  nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	r := gin.New()
	r.Use(AuthMiddleware(a))
	var gotUser models.User
	var gotClaims *utils.Claims
	r.GET("/api/protected", func(c *gin.Context) {
		u, _ := c.Get("user")
		gotUser = u.(models.User)
		cl, _ := c.Get("claims")
		gotClaims = cl.(*utils.Claims)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if gotUser.ID != 1 || gotUser.Username != "testuser" || gotUser.Nickname != "测试" {
		t.Errorf("user = %+v", gotUser)
	}
	if gotClaims == nil || gotClaims.UserID != 1 || len(gotClaims.RoleIDs) != 2 {
		t.Errorf("claims = %+v", gotClaims)
	}
}
