package controllers

import (
	"strconv"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	app *app.App
}

func NewRoleController(a *app.App) *RoleController {
	return &RoleController{app: a}
}

type getRolesQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	OrderBy  string `form:"order_by"`
}

func (ctrl *RoleController) GetRoles(c *gin.Context) {
	var req getRolesQuery
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
	filters := make(map[string]string)
	if req.OrderBy != "" {
		filters["order_by"] = req.OrderBy
	}

	roles, total, err := ctrl.app.GetRoleService().GetRoles(c, page, pageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.Success(c, gin.H{
		"data": roles,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (ctrl *RoleController) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	role, err := ctrl.app.GetRoleService().CreateRole(c, req.Name, req.Description)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "创建成功", role)
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (ctrl *RoleController) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的角色ID"))
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	role, err := ctrl.app.GetRoleService().UpdateRole(c, uint(roleID), req.Name, req.Description)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "更新成功", role)
}

func (ctrl *RoleController) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的角色ID"))
		return
	}

	if err := ctrl.app.GetRoleService().DeleteRole(c, uint(roleID)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "删除成功", nil)
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func (ctrl *RoleController) AssignPermissions(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的角色ID"))
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	if err := ctrl.app.GetRoleService().AssignPermissions(c, uint(roleID), req.PermissionIDs); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "分配权限成功", nil)
}
