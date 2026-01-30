package controllers

import (
	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	app *app.App
}

func NewAuthController(a *app.App) *AuthController {
	return &AuthController{app: a}
}

type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CaptchaID  string `json:"captcha_id" binding:"required"`
	CaptchaVal string `json:"captcha_val" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	user, token, err := ctrl.app.GetAuthService().Login(c, req.Username, req.Password, req.CaptchaID, req.CaptchaVal)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	userResponse := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"roles":    user.Roles,
	}

	c.SetCookie("token", token, 24*3600, "/", "", false, true)

	ctrl.app.Responder.Success(c, gin.H{
		"token": token,
		"user":  userResponse,
	})
}

func (ctrl *AuthController) GetCaptcha(c *gin.Context) {
	id, b64s, err := ctrl.app.GetAuthService().GenerateCaptcha(c)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.Success(c, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		ctrl.app.Responder.RespondError(c, errors.UnauthorizedMsg("未认证"))
		return
	}
	userModel := user.(models.User)

	if err := ctrl.app.GetAuthService().Logout(c, userModel.ID); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Logger().InfoContext(c, "用户退出登录", "user_id", userModel.ID, "username", userModel.Username)

	c.SetCookie("token", "", -1, "/", "", false, true)
	ctrl.app.Responder.SuccessWithMsg(c, "退出登录成功", nil)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		ctrl.app.Responder.RespondError(c, errors.UnauthorizedMsg("未认证"))
		return
	}
	userModel := user.(models.User)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	if err := ctrl.app.GetAuthService().ChangePassword(c, userModel.ID, req.OldPassword, req.NewPassword); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "密码修改成功", nil)
}

type UpdateAvatarRequest struct {
	Avatar string `json:"avatar"`
}

func (ctrl *AuthController) UpdateAvatar(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		ctrl.app.Responder.RespondError(c, errors.UnauthorizedMsg("未认证"))
		return
	}
	userModel := user.(models.User)

	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	if err := ctrl.app.GetAuthService().UpdateAvatar(c, userModel.ID, req.Avatar); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.SuccessWithMsg(c, "头像更新成功", nil)
}

func (ctrl *AuthController) GetUserPermissions(c *gin.Context) {
	claims, ok := utils.ClaimsFromContext(c)
	if !ok {
		ctrl.app.Responder.RespondError(c, errors.UnauthorizedMsg("未认证"))
		return
	}

	isSuperAdmin := claims.IsSuperAdmin
	roleIDs := claims.RoleIDs

	if isSuperAdmin {
		ctrl.app.Responder.Success(c, gin.H{
			"is_super_admin": true,
			"permissions":    []models.Permission{},
		})
		return
	}

	permissions, err := ctrl.app.GetPermissionService().GetPermissionsByRoleIDs(c, roleIDs)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	ctrl.app.Responder.Success(c, gin.H{
		"is_super_admin": false,
		"permissions":    permissions,
	})
}
