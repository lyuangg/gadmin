package services

import (
	"context"
	stderrors "errors"

	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"

	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	ctx ServiceContext
}

// NewRoleService 创建角色服务实例
func NewRoleService(ctx ServiceContext) *RoleService {
	return &RoleService{ctx: ctx}
}

// GetRoles 获取角色列表（分页和筛选）
func (s *RoleService) GetRoles(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.Role, int64, error) {
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
	var roles []models.Role

	// 构建查询
	query := s.ctx.DB().Model(&models.Role{})

	// 应用排序
	orderBy := filters["order_by"]
	if orderBy == "id" || orderBy == "id_asc" {
		query = query.Order("id ASC")
	} else if orderBy == "id_desc" {
		query = query.Order("id DESC")
	} else {
		// 默认按 id 倒序
		query = query.Order("id DESC")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Permissions").Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	// 检查角色名是否已存在
	var existingRole models.Role
	if err := s.ctx.DB().Where("name = ?", name).First(&existingRole).Error; err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.BadRequestMsg("角色名已存在")
	}

	role := models.Role{
		Name:        name,
		Description: description,
	}

	if err := s.ctx.DB().Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(ctx context.Context, roleID uint, name, description string) (*models.Role, error) {
	var role models.Role
	if err := s.ctx.DB().Where("id = ?", roleID).First(&role).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("角色不存在")
		}
		return nil, err
	}

	if name != "" {
		// 检查角色名是否与其他角色冲突
		var existingRole models.Role
		if err := s.ctx.DB().Where("name = ? AND id != ?", name, roleID).First(&existingRole).Error; err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			return nil, errors.BadRequestMsg("角色名已存在")
		}
		role.Name = name
	}

	if description != "" {
		role.Description = description
	}

	if err := s.ctx.DB().Save(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(ctx context.Context, roleID uint) error {
	var role models.Role
	if err := s.ctx.DB().Where("id = ?", roleID).First(&role).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("角色不存在")
		}
		return err
	}

	// 清除关联关系
	s.ctx.DB().Model(&role).Association("Users").Clear()
	s.ctx.DB().Model(&role).Association("Permissions").Clear()

	// 删除角色
	if err := s.ctx.DB().Delete(&role).Error; err != nil {
		return err
	}

	return nil
}

// AssignPermissions 为角色分配权限
func (s *RoleService) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	var role models.Role
	if err := s.ctx.DB().Where("id = ?", roleID).First(&role).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("角色不存在")
		}
		return err
	}

	// 查询权限
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := s.ctx.DB().Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			return err
		}
	}

	// 分配权限
	s.ctx.DB().Model(&role).Association("Permissions").Replace(permissions)

	return nil
}
