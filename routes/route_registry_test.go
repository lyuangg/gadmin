package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lyuangg/gadmin/routes/routemeta"

	"github.com/gin-gonic/gin"
)

// 使用唯一路径避免与其它测试或真实注册的路由冲突
const testRoutePath = "/admin/api/route-registry-test"

func TestRegisterRouteWithPermission_RegistersPermissionAndRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/admin/api")

	handlerCalled := false
	RegisterRouteWithPermission(group, "GET", "/route-registry-test", "测试权限", "测试分组", func(c *gin.Context) {
		handlerCalled = true
		c.String(200, "ok")
	})

	// 断言权限映射表中有完整路径与名称
	info := routemeta.GetRoutePermission("GET", testRoutePath)
	if info.Name != "测试权限" || info.Group != "测试分组" {
		t.Errorf("GetRoutePermission(GET, %s) = Name=%q Group=%q, want 测试权限 测试分组", testRoutePath, info.Name, info.Group)
	}

	// 断言路由已注册，请求可命中
	req := httptest.NewRequest(http.MethodGet, testRoutePath, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestRegisterRouteWithPermission_FullPathFromBaseAndPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/admin/api")

	// fullPath = basePath + path，path 带前导 / 时得到 /admin/api/users
	RegisterRouteWithPermission(group, "POST", "/users", "用户管理", "用户", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	info := routemeta.GetRoutePermission("POST", "/admin/api/users")
	if info.Name != "用户管理" || info.Group != "用户" {
		t.Errorf("fullPath /admin/api/users: info = %+v", info)
	}
}

func TestRegisterRouteWithPermission_MethodCase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/admin/api")

	RegisterRouteWithPermission(group, "delete", "/route-registry-delete", "删除测试", "测试", func(c *gin.Context) {
		c.Status(200)
	})

	info := routemeta.GetRoutePermission("DELETE", "/admin/api/route-registry-delete")
	if info.Name != "删除测试" {
		t.Errorf("method should be normalized: Name=%q", info.Name)
	}

	req := httptest.NewRequest(http.MethodDelete, "/admin/api/route-registry-delete", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("DELETE route should be registered: status=%d", rec.Code)
	}
}
