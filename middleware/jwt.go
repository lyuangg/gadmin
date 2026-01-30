package middleware

import (
	"net/http"
	"strings"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(a *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			cookieToken, err := c.Cookie("token")
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}

		if token == "" {
			token = c.Query("token")
		}

		redirectToLogin := func(errorMsg string, statusCode int) {
			if strings.Contains(c.GetHeader("Accept"), "text/html") {
				c.SetCookie("token", "", -1, "/", "", false, true)
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
			} else {
				if statusCode == http.StatusForbidden {
					a.Responder.RespondError(c, errors.ForbiddenMsg(errorMsg))
				} else {
					a.Responder.RespondError(c, errors.UnauthorizedMsg(errorMsg))
				}
				c.Abort()
			}
		}

		if token == "" {
			redirectToLogin("未提供认证Token", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			redirectToLogin("Token无效或已过期", http.StatusUnauthorized)
			return
		}

		// 检查 token 中的 token_version 是否匹配用户的当前 token_version
		_, tokenVersion, err := utils.GetTokenVersion(claims)
		if err != nil {
			redirectToLogin("Token格式错误", http.StatusUnauthorized)
			return
		}

		user, err := a.GetUserService().GetUserForAuth(c.Request.Context(), claims.UserID)
		if err != nil || user == nil {
			redirectToLogin("用户不存在", http.StatusUnauthorized)
			return
		}

		if tokenVersion < user.TokenVersion {
			redirectToLogin("Token已失效", http.StatusUnauthorized)
			return
		}

		// 检查用户状态（0=禁用，1=启用）
		if user.Status == 0 {
			redirectToLogin("用户已被禁用", http.StatusForbidden)
			return
		}

		user.ID = claims.UserID
		user.Username = claims.Username
		user.Nickname = claims.Nickname
		user.Type = claims.Type

		user.Roles = make([]models.Role, len(claims.RoleIDs))
		for i, roleID := range claims.RoleIDs {
			user.Roles[i] = models.Role{ID: roleID}
		}

		c.Set("user", *user)
		c.Set("claims", claims)

		c.Next()
	}
}
