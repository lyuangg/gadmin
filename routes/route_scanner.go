package routes

import (
	"context"
	"strings"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/routes/routemeta"

	"github.com/gin-gonic/gin"
)

type RouteScanner struct {
	router *gin.Engine
	app    *app.App
}

func NewRouteScanner(router *gin.Engine, a *app.App) *RouteScanner {
	return &RouteScanner{router: router, app: a}
}

func (rs *RouteScanner) ScanAndImport() error {
	routes := rs.router.Routes()

	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/admin/api/") {
			continue
		}

		permissionInfo := routemeta.GetRoutePermission(route.Method, route.Path)
		if permissionInfo.Name == "" {
			continue
		}

		var permission models.Permission
		result := rs.app.DB().Where("path = ? AND method = ?", route.Path, route.Method).First(&permission)

		if result.Error != nil {
			permission = models.Permission{
				Path:       route.Path,
				Method:     route.Method,
				Name:       permissionInfo.Name,
				Group:      permissionInfo.Group,
				AutoImport: true,
			}

			if err := rs.app.DB().Create(&permission).Error; err != nil {
				rs.app.Logger().ErrorContext(context.Background(), "导入权限失败",
					"method", route.Method,
					"path", route.Path,
					"error", err)
				continue
			}

			rs.app.Logger().InfoContext(context.Background(), "自动导入权限",
				"method", route.Method,
				"path", route.Path,
				"name", permission.Name,
				"group", permission.Group)
		} else {
			needUpdate := false
			if permission.Name != permissionInfo.Name {
				permission.Name = permissionInfo.Name
				needUpdate = true
			}
			if permission.Group != permissionInfo.Group {
				permission.Group = permissionInfo.Group
				needUpdate = true
			}
			if !permission.AutoImport {
				permission.AutoImport = true
				needUpdate = true
			}

			if needUpdate {
				rs.app.DB().Save(&permission)
				rs.app.Logger().InfoContext(context.Background(), "更新权限信息",
					"method", route.Method,
					"path", route.Path,
					"name", permission.Name,
					"group", permission.Group)
			}
		}
	}

	return nil
}
