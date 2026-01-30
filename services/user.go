package services

import (
	"context"
	stderrors "errors"
	"math/rand"

	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	ctx ServiceContext
}

func NewUserService(ctx ServiceContext) *UserService {
	return &UserService{ctx: ctx}
}

func (s *UserService) GetUsers(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	var users []models.User

	query := s.ctx.DB().Model(&models.User{})
	if username, ok := filters["username"]; ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if nickname, ok := filters["nickname"]; ok && nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if typeStr, ok := filters["type"]; ok && typeStr != "" {
		query = query.Where("type = ?", typeStr)
	}
	if statusStr, ok := filters["status"]; ok && statusStr != "" {
		query = query.Where("status = ?", statusStr)
	}
	if roleID, ok := filters["role_id"]; ok && roleID != "" {
		subQuery := s.ctx.DB().Model(&models.UserRole{}).Select("user_id").Where("role_id = ?", roleID)
		query = query.Where("id IN (?)", subQuery)
	}

	switch orderBy := filters["order_by"]; orderBy {
	case "id", "id_asc":
		query = query.Order("id ASC")
	case "id_desc":
		query = query.Order("id DESC")
	default:
		query = query.Order("id DESC")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserForAuth 供认证中间件使用，仅查询校验 token_version 与状态所需字段
func (s *UserService) GetUserForAuth(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := s.ctx.DB().Select("id", "username", "nickname", "type", "status", "token_version").
		Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CreateUser(ctx context.Context, username, password, nickname string, userType int, remark string, roleIDs []uint) (*models.User, error) {
	var existingUser models.User
	if err := s.ctx.DB().Where("username = ?", username).First(&existingUser).Error; err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.BadRequestMsg("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.InternalErrorMsg("密码加密失败")
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: nickname,
		Type:     userType,
		Status:   1,
		Remark:   remark,
	}

	if err := s.ctx.DB().Create(&user).Error; err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := s.ctx.DB().Where("id IN ?", roleIDs).Find(&roles).Error; err == nil {
			s.ctx.DB().Model(&user).Association("Roles").Replace(roles)
		}
	}

	return &user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, nickname, password, remark string, roleIDs []uint) error {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}

	if nickname != "" {
		user.Nickname = nickname
	}
	user.Remark = remark

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return errors.InternalErrorMsg("密码加密失败")
		}
		user.Password = string(hashedPassword)
	}

	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return errors.InternalErrorMsg("更新用户失败")
	}

	if roleIDs != nil {
		var roles []models.Role
		if len(roleIDs) > 0 {
			if err := s.ctx.DB().Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
		}
		s.ctx.DB().Model(&user).Association("Roles").Replace(roles)
	}

	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, userID uint) (string, error) {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.NotFoundMsg("用户不存在")
		}
		return "", err
	}

	newPassword := s.generateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.InternalErrorMsg("密码加密失败")
	}

	user.Password = string(hashedPassword)
	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return "", err
	}

	return newPassword, nil
}

func (s *UserService) ToggleStatus(ctx context.Context, userID uint) error {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}

	if user.Status == 1 {
		user.Status = 0
	} else {
		user.Status = 1
	}

	if err := s.ctx.DB().Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) generateRandomPassword() string {
	const numbers = "0123456789"
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const allChars = numbers + letters

	password := make([]byte, 6)
	password[0] = numbers[rand.Intn(len(numbers))]
	password[1] = letters[rand.Intn(len(letters))]
	for i := 2; i < 6; i++ {
		password[i] = allChars[rand.Intn(len(allChars))]
	}
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password)
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint) error {
	var user models.User
	if err := s.ctx.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("用户不存在")
		}
		return err
	}

	s.ctx.DB().Model(&user).Association("Roles").Clear()
	if err := s.ctx.DB().Delete(&user).Error; err != nil {
		return err
	}

	return nil
}
