package middleware

import (
	"regexp"
	"strings"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(a *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := utils.ClaimsFromContext(c)
		if !ok {
			a.Responder.RespondError(c, errors.UnauthorizedMsg("未认证"))
			c.Abort()
			return
		}

		isSuperAdmin := claims.IsSuperAdmin
		roleIDs := claims.RoleIDs

		if isSuperAdmin {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		var permissions []models.Permission
		if len(roleIDs) > 0 {
			var err error
			permissions, err = a.GetPermissionService().GetPermissionsByRoleIDs(c, roleIDs)
			if err != nil {
				a.Responder.RespondError(c, errors.InternalErrorMsg("查询权限失败"))
				c.Abort()
				return
			}
		}

		hasPermission := false
		for _, perm := range permissions {
			if matchPermission(perm.Path, perm.Method, path, method) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			a.Responder.RespondError(c, errors.ForbiddenMsg("没有权限访问此资源"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func matchPermission(permPath, permMethod, reqPath, reqMethod string) bool {
	if !strings.EqualFold(permMethod, reqMethod) {
		return false
	}
	if permPath == reqPath {
		return true
	}
	if strings.HasSuffix(permPath, "/*") {
		prefix := strings.TrimSuffix(permPath, "/*")
		if strings.HasPrefix(reqPath, prefix+"/") {
			return true
		}
	}

	// :id 等路径参数按正则匹配
	pattern := regexp.QuoteMeta(permPath)
	pattern = regexp.MustCompile(`:\\w+`).ReplaceAllString(pattern, `[^/]+`)
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return false
	}
	if re.MatchString(reqPath) {
		return true
	}

	return false
}
