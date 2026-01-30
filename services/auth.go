package services

import (
	"context"
	stderrors "errors"

	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	ctx ServiceContext
}

func NewAuthService(ctx ServiceContext) *AuthService {
	return &AuthService{ctx: ctx}
}

func (s *AuthService) Login(ctx context.Context, username, password, captchaID, captchaVal string) (*models.User, string, error) {
	if !s.ctx.GetCaptchaProvider().Verify(captchaID, captchaVal) {
		return nil, "", errors.UnauthorizedMsg("验证码错误")
	}

	var user models.User
	if err := s.ctx.DB().Where("username = ?", username).Preload("Roles").First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.UnauthorizedMsg("用户名或密码错误")
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.UnauthorizedMsg("用户名或密码错误")
	}

	isSuperAdmin := false
	var roleIDs []uint
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
		if role.Name == "超级管理员" {
			isSuperAdmin = true
		}
	}

	token, err := s.ctx.GetTokenGenerator().GenerateToken(
		user.ID,
		user.Username,
		user.Nickname,
		user.Type,
		isSuperAdmin,
		roleIDs,
		user.TokenVersion,
	)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *AuthService) GenerateCaptcha(ctx context.Context) (string, string, error) {
	return s.ctx.GetCaptchaProvider().Generate(ctx)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.UnauthorizedMsg("当前密码错误")
	}

	if len(newPassword) < 6 {
		return errors.BadRequestMsg("新密码长度不能少于6位")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalErrorMsg("密码加密失败")
	}

	user.Password = string(hashedPassword)
	user.TokenVersion++
	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// UpdateAvatar 更新当前用户头像
func (s *AuthService) UpdateAvatar(ctx context.Context, userID uint, avatarURL string) error {
	// 查找用户
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}

	user.Avatar = avatarURL
	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// Logout 递增用户 token_version，使已签发的 token 失效
func (s *AuthService) Logout(ctx context.Context, userID uint) error {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}
	user.TokenVersion++
	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return errors.InternalErrorMsg("更新Token版本失败")
	}
	return nil
}
