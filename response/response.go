package response

import (
	stderrors "errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type IResponder interface {
	Success(c *gin.Context, data interface{})
	SuccessWithMsg(c *gin.Context, msg string, data interface{})
	RespondError(c *gin.Context, err error)
	Error(c *gin.Context, code int, msg string)
}

// Response 统一响应结构（JSON 体）
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type Responder struct {
	logger *slog.Logger
}

// NewResponder 创建 Responder，logger 由 app 在初始化时传入
func NewResponder(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

func (r *Responder) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func (r *Responder) SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: msg, Data: data})
}

func (r *Responder) Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil})
}

// RespondError BizError 用其 code/msg；非 BizError 打日志并返回 500
func (r *Responder) RespondError(c *gin.Context, err error) {
	var biz *errors.BizError
	if stderrors.As(err, &biz) {
		r.Error(c, biz.Code, biz.Msg)
		return
	}
	stack := string(debug.Stack())
	attrs := []any{
		"err", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"client_ip", c.ClientIP(),
		"stack", stack,
	}
	if c.Request.URL.RawQuery != "" {
		attrs = append(attrs, "query", c.Request.URL.RawQuery)
	}
	r.logger.ErrorContext(c, "internal error: "+err.Error(), attrs...)
	r.Error(c, errors.CodeInternalError, "服务器内部错误")
}
