package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
)

const traceIDContextKey = "trace_id"
const traceIDHeader = "X-Trace-Id"

func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(traceIDHeader)
		if traceID == "" {
			traceID = generateTraceID()
		}
		c.Set(traceIDContextKey, traceID)
		ctx := context.WithValue(c.Request.Context(), traceIDContextKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(traceIDHeader, traceID)
		c.Next()
	}
}

// generateTraceID 生成 32 位十六进制 trace id
func generateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// 降级：用纳秒时间戳填充 16 字节再转 hex（rand 失败极少发生）
		ns := time.Now().UnixNano()
		for i := 0; i < 16; i++ {
			b[i] = byte(ns >> (i * 8))
		}
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b)
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val := ctx.Value(traceIDContextKey)
	if val == nil {
		return "", false
	}
	s, ok := val.(string)
	return s, ok && s != ""
}
