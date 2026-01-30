package routemeta

import (
	"testing"
)

// resetRoutePermissionMapForTest 仅用于单测，清空路由权限映射表（与 permission.go 同包可访问未导出变量）
func resetRoutePermissionMapForTest() {
	routePermissionMapMutex.Lock()
	defer routePermissionMapMutex.Unlock()
	routePermissionMap = make(map[string]RoutePermissionInfo)
}

func TestGetRoutePermission_ExactMatch(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/users", "用户列表", "用户管理")

	info := GetRoutePermission("GET", "/admin/api/users")
	if info.Name != "用户列表" || info.Group != "用户管理" {
		t.Errorf("GetRoutePermission(GET, /admin/api/users) = %+v", info)
	}
}

func TestGetRoutePermission_MethodCaseInsensitive(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("get", "/admin/api/roles", "角色列表", "角色管理")

	info := GetRoutePermission("GET", "/admin/api/roles")
	if info.Name != "角色列表" {
		t.Errorf("method should be case-insensitive, got Name=%q", info.Name)
	}
}

func TestGetRoutePermission_ParamMatch(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/roles/:id/permissions", "角色权限列表", "角色管理")

	info := GetRoutePermission("GET", "/admin/api/roles/1/permissions")
	if info.Name != "角色权限列表" || info.Group != "角色管理" {
		t.Errorf("GetRoutePermission(GET, /admin/api/roles/1/permissions) = %+v", info)
	}
}

func TestGetRoutePermission_ParamMatchSingleSegment(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/users/:id", "用户详情", "用户管理")

	info := GetRoutePermission("GET", "/admin/api/users/42")
	if info.Name != "用户详情" {
		t.Errorf("GetRoutePermission(GET, /admin/api/users/42) = %+v", info)
	}
}

func TestGetRoutePermission_NoMatchReturnsEmpty(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/users", "用户列表", "用户管理")

	info := GetRoutePermission("GET", "/admin/api/unknown")
	if info.Name != "" || info.Group != "" {
		t.Errorf("unregistered path should return empty, got %+v", info)
	}
}

func TestGetRoutePermission_MethodMismatch(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/users", "用户列表", "用户管理")

	info := GetRoutePermission("POST", "/admin/api/users")
	if info.Name != "" {
		t.Errorf("method mismatch should return empty, got Name=%q", info.Name)
	}
}

func TestGetRoutePermission_ParamSegmentCountMismatch(t *testing.T) {
	resetRoutePermissionMapForTest()
	RegisterRoutePermission("GET", "/admin/api/roles/:id", "角色详情", "角色管理")

	// 实际路径多一段，不应匹配
	info := GetRoutePermission("GET", "/admin/api/roles/1/extra")
	if info.Name != "" {
		t.Errorf("segment count mismatch should not match, got Name=%q", info.Name)
	}
}
