package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/routes/routemeta"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

// 操作记录的最大数据大小限制（50KB，MySQL TEXT 类型最大为 64KB）
const maxOperationLogSize = 50 * 1024

// OperationLogMiddleware 记录 /admin/api 下 PUT、DELETE、POST 的请求与响应
func OperationLogMiddleware(a *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/admin/api/") {
			c.Next()
			return
		}

		method := c.Request.Method
		if method != "PUT" && method != "DELETE" && method != "POST" {
			c.Next()
			return
		}

		startTime := time.Now()

		var userID uint
		var username string
		if claims, ok := utils.ClaimsFromContext(c); ok {
			userID = claims.UserID
			username = claims.Username
		}

		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if len(bodyBytes) > maxOperationLogSize {
					requestBody = string(bodyBytes[:maxOperationLogSize]) + "...(truncated)"
				} else {
					requestBody = string(bodyBytes)
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = blw

		c.Next()

		duration := time.Since(startTime).Milliseconds()
		responseBody := blw.body.String()
		if len(responseBody) > maxOperationLogSize {
			responseBody = responseBody[:maxOperationLogSize] + "...(truncated)"
		}

		formattedRequest := formatJSON(requestBody)
		formattedResponse := formatJSON(responseBody)
		routePermissionInfo := routemeta.GetRoutePermission(method, c.Request.URL.Path)
		routeName := routePermissionInfo.Name

		operationLog := models.OperationLog{
			UserID:     userID,
			Username:   username,
			Method:     method,
			Path:       c.Request.URL.Path,
			RouteName:  routeName,
			Request:    formattedRequest,
			Response:   formattedResponse,
			StatusCode: c.Writer.Status(),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Duration:   duration,
		}

		reqCtx := c.Copy()
		go func() {
			if err := a.DB().Create(&operationLog).Error; err != nil {
				a.Logger().ErrorContext(reqCtx, "保存操作记录失败",
					"error", err,
					"path", reqCtx.Request.URL.Path,
					"method", method)
			}
		}()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func formatJSON(s string) string {
	if s == "" {
		return ""
	}
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(s), &jsonObj); err != nil {
		return s
	}
	formatted, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return s
	}

	return string(formatted)
}
