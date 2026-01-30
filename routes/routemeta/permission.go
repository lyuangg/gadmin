package routemeta

import (
	"fmt"
	"strings"
	"sync"
)

// RoutePermissionInfo 路由权限信息
type RoutePermissionInfo struct {
	Name  string // 权限名称
	Group string // 权限分组名称
}

// routePermissionMap 存储路由路径+方法与权限信息的映射
// key 格式: "METHOD:/path" (例如: "GET:/admin/api/users")
// value: 权限信息（包含名称和分组）
var routePermissionMap = make(map[string]RoutePermissionInfo)
var routePermissionMapMutex sync.RWMutex

// RegisterRoutePermission 注册路由权限信息到映射表
func RegisterRoutePermission(method, path, permissionName, groupName string) {
	key := getRouteKey(method, path)
	routePermissionMapMutex.Lock()
	defer routePermissionMapMutex.Unlock()
	routePermissionMap[key] = RoutePermissionInfo{
		Name:  permissionName,
		Group: groupName,
	}
}

// GetRoutePermission 根据路由路径和方法获取配置的权限信息
// 支持路由参数匹配，例如 /users/123 可以匹配到 /users/:id
// 如果未配置，返回空信息
func GetRoutePermission(method, path string) RoutePermissionInfo {
	key := getRouteKey(method, path)
	routePermissionMapMutex.RLock()
	defer routePermissionMapMutex.RUnlock()

	// 首先尝试精确匹配
	if info, exists := routePermissionMap[key]; exists {
		return info
	}

	// 如果精确匹配失败，尝试匹配带参数的路由
	for routeKey, info := range routePermissionMap {
		parts := strings.SplitN(routeKey, ":", 2)
		if len(parts) != 2 {
			continue
		}
		routeMethod := parts[0]
		routePath := parts[1]
		if !strings.EqualFold(routeMethod, method) {
			continue
		}
		if matchRoutePath(routePath, path) {
			return info
		}
	}

	return RoutePermissionInfo{}
}

// matchRoutePath 匹配路由路径，支持路由参数
// 例如：/users/:id 可以匹配 /users/123
func matchRoutePath(routePath, actualPath string) bool {
	routeParts := strings.Split(routePath, "/")
	actualParts := strings.Split(actualPath, "/")
	if len(routeParts) != len(actualParts) {
		return false
	}
	for i := 0; i < len(routeParts); i++ {
		routePart := routeParts[i]
		actualPart := actualParts[i]
		if strings.HasPrefix(routePart, ":") {
			continue
		}
		if routePart != actualPart {
			return false
		}
	}
	return true
}

func getRouteKey(method, path string) string {
	return fmt.Sprintf("%s:%s", strings.ToUpper(method), path)
}
