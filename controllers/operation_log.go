package controllers

import (
	stderrors "errors"
	"time"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct {
	app *app.App
}

func NewOperationLogController(a *app.App) *OperationLogController {
	return &OperationLogController{app: a}
}

type getOperationLogsQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	Username   string `form:"username"`
	Method     string `form:"method"`
	Path       string `form:"path"`
	StatusCode string `form:"status_code"`
	OrderBy    string `form:"order_by"`
}

func (ctrl *OperationLogController) GetOperationLogs(c *gin.Context) {
	var req getOperationLogsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	if req.StartTime != "" {
		if _, err := time.Parse(time.RFC3339, req.StartTime); err != nil {
			ctrl.app.Responder.RespondError(c, errors.BadRequestErr(stderrors.New("start_time 格式错误，需使用 RFC3339，例如 2025-01-01T00:00:00Z")))
			return
		}
	}
	if req.EndTime != "" {
		if _, err := time.Parse(time.RFC3339, req.EndTime); err != nil {
			ctrl.app.Responder.RespondError(c, errors.BadRequestErr(stderrors.New("end_time 格式错误，需使用 RFC3339，例如 2025-01-01T23:59:59Z")))
			return
		}
	}

	filters := map[string]string{
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
		"username":    req.Username,
		"method":      req.Method,
		"path":        req.Path,
		"status_code": req.StatusCode,
		"order_by":    req.OrderBy,
	}

	logs, total, err := ctrl.app.GetOperationLogService().GetOperationLogs(c, page, pageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.Success(c, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}
