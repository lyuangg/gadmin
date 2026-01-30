package services

import (
	"context"
	stderrors "errors"

	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"

	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	ctx ServiceContext
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(ctx ServiceContext) *PermissionService {
	return &PermissionService{ctx: ctx}
}

// GetPermissionsByRoleIDs 根据角色 ID 列表查询合并去重后的权限（使用模型 + Preload，表名由 NamingStrategy 统一处理）
func (s *PermissionService) GetPermissionsByRoleIDs(ctx context.Context, roleIDs []uint) ([]models.Permission, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var roles []models.Role
	if err := s.ctx.DB().Where("id IN ?", roleIDs).Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	permMap := make(map[uint]models.Permission)
	for _, r := range roles {
		for _, p := range r.Permissions {
			permMap[p.ID] = p
		}
	}
	permissions := make([]models.Permission, 0, len(permMap))
	for _, p := range permMap {
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// GetPermissions 获取权限列表（分页和筛选）
func (s *PermissionService) GetPermissions(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.Permission, int64, error) {
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
	var permissions []models.Permission

	// 构建查询
	query := s.ctx.DB().Model(&models.Permission{})

	// 应用筛选条件
	if path, ok := filters["path"]; ok && path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if method, ok := filters["method"]; ok && method != "" {
		query = query.Where("method = ?", method)
	}
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if group, ok := filters["group"]; ok && group != "" {
		query = query.Where("`group` LIKE ?", "%"+group+"%")
	}

	// 应用排序
	orderBy := filters["order_by"]
	switch orderBy {
	case "id", "id_asc":
		query = query.Order("id ASC")
	case "id_desc":
		query = query.Order("id DESC")
	default:
		// 默认按 id 倒序
		query = query.Order("id DESC")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(ctx context.Context, path, method, name, group, description string) (*models.Permission, error) {
	// 检查权限是否已存在
	var existingPermission models.Permission
	if err := s.ctx.DB().Where("path = ? AND method = ?", path, method).First(&existingPermission).Error; err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.BadRequestMsg("权限已存在")
	}

	permission := models.Permission{
		Path:        path,
		Method:      method,
		Name:        name,
		Group:       group,
		Description: description,
		AutoImport:  false,
	}

	if err := s.ctx.DB().Create(&permission).Error; err != nil {
		return nil, err
	}

	return &permission, nil
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(ctx context.Context, permissionID uint, name, group, description string) (*models.Permission, error) {
	var permission models.Permission
	if err := s.ctx.DB().Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("权限不存在")
		}
		return nil, err
	}

	if name != "" {
		permission.Name = name
	}

	if group != "" {
		permission.Group = group
	}

	if description != "" {
		permission.Description = description
	}

	if err := s.ctx.DB().Save(&permission).Error; err != nil {
		return nil, err
	}

	return &permission, nil
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(ctx context.Context, permissionID uint) error {
	var permission models.Permission
	if err := s.ctx.DB().Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("权限不存在")
		}
		return err
	}

	// 清除关联关系
	s.ctx.DB().Model(&permission).Association("Roles").Clear()

	// 删除权限
	if err := s.ctx.DB().Delete(&permission).Error; err != nil {
		return err
	}

	return nil
}

// BatchDeletePermissions 批量删除权限
func (s *PermissionService) BatchDeletePermissions(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return errors.BadRequestMsg("权限ID列表不能为空")
	}

	// 检查所有权限是否存在
	var count int64
	if err := s.ctx.DB().Model(&models.Permission{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}

	if int(count) != len(ids) {
		return errors.BadRequestMsg("部分权限不存在")
	}

	// 批量清除关联关系 - 使用模型关联，让 GORM 自动应用表前缀
	var permissions []models.Permission
	if err := s.ctx.DB().Where("id IN ?", ids).Find(&permissions).Error; err != nil {
		return err
	}
	for _, permission := range permissions {
		if err := s.ctx.DB().Model(&permission).Association("Roles").Clear(); err != nil {
			return err
		}
	}

	// 批量删除权限
	if err := s.ctx.DB().Where("id IN ?", ids).Delete(&models.Permission{}).Error; err != nil {
		return err
	}

	return nil
}
