package controllers

import (
	"strconv"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	app *app.App
}

func NewPermissionController(a *app.App) *PermissionController {
	return &PermissionController{app: a}
}

type getPermissionsQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Path     string `form:"path"`
	Method   string `form:"method"`
	Name     string `form:"name"`
	Group    string `form:"group"`
	OrderBy  string `form:"order_by"`
}

func (ctrl *PermissionController) GetPermissions(c *gin.Context) {
	var req getPermissionsQuery
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
	filters := map[string]string{
		"path":     req.Path,
		"method":   req.Method,
		"name":     req.Name,
		"group":    req.Group,
		"order_by": req.OrderBy,
	}

	permissions, total, err := ctrl.app.GetPermissionService().GetPermissions(c, page, pageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.Success(c, gin.H{
		"data": permissions,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

type CreatePermissionRequest struct {
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

func (ctrl *PermissionController) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	permission, err := ctrl.app.GetPermissionService().CreatePermission(c, req.Path, req.Method, req.Name, req.Group, req.Description)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "创建成功", permission)
}

type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

func (ctrl *PermissionController) UpdatePermission(c *gin.Context) {
	permissionIDStr := c.Param("id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的权限ID"))
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	permission, err := ctrl.app.GetPermissionService().UpdatePermission(c, uint(permissionID), req.Name, req.Group, req.Description)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "更新成功", permission)
}

func (ctrl *PermissionController) DeletePermission(c *gin.Context) {
	permissionIDStr := c.Param("id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的权限ID"))
		return
	}

	if err := ctrl.app.GetPermissionService().DeletePermission(c, uint(permissionID)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "删除成功", nil)
}

type BatchDeletePermissionsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

func (ctrl *PermissionController) BatchDeletePermissions(c *gin.Context) {
	var req BatchDeletePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	if err := ctrl.app.GetPermissionService().BatchDeletePermissions(c, req.IDs); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "批量删除成功", nil)
}
