package middleware

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/lyuangg/gadmin/app"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware 记录请求；API 请求额外记录请求体/响应体
func LoggingMiddleware(a *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		isAPI := strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/admin/api/")

		var reqBody string
		if isAPI && c.Request.Body != nil {
			if bodyBytes, err := io.ReadAll(c.Request.Body); err == nil {
				reqBody = string(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		var blw *bodyLogWriter
		if isAPI {
			blw = &bodyLogWriter{
				ResponseWriter: c.Writer,
				body:           &bytes.Buffer{},
			}
			c.Writer = blw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", latency.Milliseconds(),
			"client_ip", clientIP,
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}

		body := method + " " + path + " " + strconv.Itoa(status)
		if isAPI && blw != nil {
			respBody := blw.body.String()
			attrs = append(attrs, "req_body", reqBody, "resp_body", respBody)
			a.Logger().InfoContext(c, "[API] "+body, attrs...)
		} else {
			a.Logger().InfoContext(c, "[PAGE] "+body, attrs...)
		}
	}
}
