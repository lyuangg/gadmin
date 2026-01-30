package services

import (
	"context"

	"github.com/lyuangg/gadmin/models"
)

type IAuthService interface {
	Login(ctx context.Context, username, password, captchaID, captchaVal string) (*models.User, string, error)
	GenerateCaptcha(ctx context.Context) (string, string, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	UpdateAvatar(ctx context.Context, userID uint, avatarURL string) error
	Logout(ctx context.Context, userID uint) error
}

type IUserService interface {
	GetUsers(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.User, int64, error)
	GetUserForAuth(ctx context.Context, userID uint) (*models.User, error) // 认证中间件用：按 ID 查用户（id, username, nickname, type, status, token_version）
	CreateUser(ctx context.Context, username, password, nickname string, userType int, remark string, roleIDs []uint) (*models.User, error)
	UpdateUser(ctx context.Context, userID uint, nickname, password, remark string, roleIDs []uint) error
	DeleteUser(ctx context.Context, userID uint) error
	ResetPassword(ctx context.Context, userID uint) (string, error)
	ToggleStatus(ctx context.Context, userID uint) error
}

type IRoleService interface {
	GetRoles(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.Role, int64, error)
	CreateRole(ctx context.Context, name, description string) (*models.Role, error)
	UpdateRole(ctx context.Context, roleID uint, name, description string) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID uint) error
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

type IPermissionService interface {
	GetPermissions(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.Permission, int64, error)
	CreatePermission(ctx context.Context, path, method, name, group, description string) (*models.Permission, error)
	UpdatePermission(ctx context.Context, permissionID uint, name, group, description string) (*models.Permission, error)
	DeletePermission(ctx context.Context, permissionID uint) error
	BatchDeletePermissions(ctx context.Context, ids []uint) error
	GetPermissionsByRoleIDs(ctx context.Context, roleIDs []uint) ([]models.Permission, error)
}

type IOperationLogService interface {
	GetOperationLogs(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.OperationLog, int64, error)
	CleanOldLogs(ctx context.Context, retain int) (int64, error)
}

type IDictionaryService interface {
	GetTypes(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.DictType, int64, error)
	CreateType(ctx context.Context, code, name, remark string) (*models.DictType, error)
	UpdateType(ctx context.Context, id uint, code, name, remark string) (*models.DictType, error)
	DeleteType(ctx context.Context, id uint) error
	GetItems(ctx context.Context, typeID uint, typeCode string, page, pageSize int, filters map[string]string) ([]models.DictItem, int64, error)
	GetItemsByCode(ctx context.Context, typeCode string) ([]models.DictItem, error)
	CreateItem(ctx context.Context, typeID uint, label, value string, sort int, status int, remark string) (*models.DictItem, error)
	UpdateItem(ctx context.Context, id uint, label, value string, sort *int, status *int, remark string) (*models.DictItem, error)
	DeleteItem(ctx context.Context, id uint) error
}
