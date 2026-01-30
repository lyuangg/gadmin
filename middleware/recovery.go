package middleware

import (
	"runtime/debug"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 捕获 panic 并返回统一 JSON 错误
func RecoveryMiddleware(a *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				a.Logger().ErrorContext(c, "panic recovered",
					"panic", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"client_ip", c.ClientIP(),
					"stack", stack,
				)
				a.Responder.Error(c, errors.CodeInternalError, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
