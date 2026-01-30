package controllers

import (
	"strconv"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	app *app.App
}

func NewUserController(a *app.App) *UserController {
	return &UserController{app: a}
}

type getUsersQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Username string `form:"username"`
	Nickname string `form:"nickname"`
	Type     string `form:"type"`
	Status   string `form:"status"`
	RoleID   string `form:"role_id"`
	OrderBy  string `form:"order_by"`
}

func (ctrl *UserController) GetUsers(c *gin.Context) {
	var req getUsersQuery
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
		"username": req.Username,
		"nickname": req.Nickname,
		"type":     req.Type,
		"status":   req.Status,
		"role_id":  req.RoleID,
		"order_by": req.OrderBy,
	}

	users, total, err := ctrl.app.GetUserService().GetUsers(c, page, pageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	var userList []gin.H
	for _, user := range users {
		userList = append(userList, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"type":       user.Type,
			"status":     user.Status,
			"remark":     user.Remark,
			"roles":      user.Roles,
			"created_at": user.CreatedAt,
		})
	}

	ctrl.app.Responder.Success(c, gin.H{
		"data": userList,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Type     int    `json:"type"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"role_ids"`
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	user, err := ctrl.app.GetUserService().CreateUser(c, req.Username, req.Password, req.Nickname, req.Type, req.Remark, req.RoleIDs)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "创建成功", gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"role_ids"`
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的用户ID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	if err := ctrl.app.GetUserService().UpdateUser(c, uint(userID), req.Nickname, req.Password, req.Remark, req.RoleIDs); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "更新成功", nil)
}

func (ctrl *UserController) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的用户ID"))
		return
	}

	if err := ctrl.app.GetUserService().DeleteUser(c, uint(userID)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "删除成功", nil)
}

func (ctrl *UserController) ResetPassword(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的用户ID"))
		return
	}

	newPassword, err := ctrl.app.GetUserService().ResetPassword(c, uint(userID))
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "密码重置成功", gin.H{
		"password": newPassword,
	})
}

func (ctrl *UserController) ToggleStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的用户ID"))
		return
	}

	if err := ctrl.app.GetUserService().ToggleStatus(c, uint(userID)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "状态更新成功", nil)
}
