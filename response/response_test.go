package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
	"log/slog"
)

func TestResponder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewResponder(slog.Default())
	router := gin.New()
	router.GET("/ok", func(c *gin.Context) {
		r.Success(c, gin.H{"id": 1})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body Response
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != 0 || body.Msg != "success" {
		t.Errorf("Code=%d Msg=%q", body.Code, body.Msg)
	}
	if body.Data == nil {
		t.Fatal("Data should not be nil")
	}
}

func TestResponder_SuccessWithMsg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewResponder(slog.Default())
	router := gin.New()
	router.GET("/msg", func(c *gin.Context) {
		r.SuccessWithMsg(c, "创建成功", gin.H{"id": 2})
	})

	req := httptest.NewRequest(http.MethodGet, "/msg", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body Response
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != 0 || body.Msg != "创建成功" {
		t.Errorf("Code=%d Msg=%q", body.Code, body.Msg)
	}
}

func TestResponder_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewResponder(slog.Default())
	router := gin.New()
	router.GET("/err", func(c *gin.Context) {
		r.Error(c, errors.CodeBadRequest, "参数错误")
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d (Responder.Error 写 200+JSON)", rec.Code)
	}
	var body Response
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != errors.CodeBadRequest || body.Msg != "参数错误" {
		t.Errorf("Code=%d Msg=%q", body.Code, body.Msg)
	}
	if body.Data != nil {
		t.Error("Error response Data should be nil")
	}
}

func TestResponder_RespondError_BizError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewResponder(slog.Default())
	router := gin.New()
	router.GET("/bizerr", func(c *gin.Context) {
		r.RespondError(c, errors.UnauthorizedMsg("Token无效"))
	})

	req := httptest.NewRequest(http.MethodGet, "/bizerr", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body Response
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != errors.CodeUnauthorized || body.Msg != "Token无效" {
		t.Errorf("BizError should use its code/msg: Code=%d Msg=%q", body.Code, body.Msg)
	}
}

func TestResponder_RespondError_NonBizError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewResponder(slog.Default())
	router := gin.New()
	router.GET("/internal", func(c *gin.Context) {
		r.RespondError(c, http.ErrNotSupported)
	})

	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body Response
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Code != errors.CodeInternalError || body.Msg != "服务器内部错误" {
		t.Errorf("non-BizError should return 500: Code=%d Msg=%q", body.Code, body.Msg)
	}
}
