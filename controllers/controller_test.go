package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

// newGinContext 创建用于单测的 gin.Context（POST JSON body）
func newGinContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(method, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// newGinContextWithUser 创建带当前用户的 context（用于需认证的 handler）
func newGinContextWithUser(method, path string, body []byte, user models.User) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newGinContext(method, path, body)
	c.Set("user", user)
	return c, w
}

// newGinContextWithClaims 创建带 claims 的 context（用于 GetUserPermissions 等）
func newGinContextWithClaims(method, path string, body []byte, claims *utils.Claims) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newGinContext(method, path, body)
	c.Set("claims", claims)
	return c, w
}

// newGinContextWithParam 创建带路径参数 id 的 context
func newGinContextWithParam(method, path string, body []byte, paramKey, paramValue string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newGinContext(method, path, body)
	c.Params = gin.Params{{Key: paramKey, Value: paramValue}}
	return c, w
}

func newGinContextGET(path string) (*gin.Context, *httptest.ResponseRecorder) {
	return newGinContext(http.MethodGet, path, nil)
}
