package routes

import (
	"strings"

	"github.com/lyuangg/gadmin/routes/routemeta"

	"github.com/gin-gonic/gin"
)

// RegisterRouteWithPermission 注册路由并写入权限映射表（供扫描导入）
func RegisterRouteWithPermission(group gin.IRoutes, method, path, permissionName, groupName string, handler gin.HandlerFunc) {
	var basePath string
	if routerGroup, ok := group.(*gin.RouterGroup); ok {
		basePath = routerGroup.BasePath()
	}
	fullPath := basePath + path
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}
	routemeta.RegisterRoutePermission(method, fullPath, permissionName, groupName)

	switch strings.ToUpper(method) {
	case "GET":
		group.GET(path, handler)
	case "POST":
		group.POST(path, handler)
	case "PUT":
		group.PUT(path, handler)
	case "PATCH":
		group.PATCH(path, handler)
	case "DELETE":
		group.DELETE(path, handler)
	default:
		group.Handle(method, path, handler)
	}
}
