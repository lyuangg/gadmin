package services

import (
	"context"

	"github.com/lyuangg/gadmin/models"
)

// FakeAuthService 单测用 IAuthService mock，可配置各方法返回值
type FakeAuthService struct {
	LoginUser  *models.User
	LoginToken string
	LoginErr   error

	GenerateID  string
	GenerateB64 string
	GenerateErr error

	ChangePasswordErr error
	UpdateAvatarErr   error
	LogoutErr         error
}

func (f *FakeAuthService) Login(_ context.Context, _, _, _, _ string) (*models.User, string, error) {
	return f.LoginUser, f.LoginToken, f.LoginErr
}
func (f *FakeAuthService) GenerateCaptcha(_ context.Context) (string, string, error) {
	return f.GenerateID, f.GenerateB64, f.GenerateErr
}
func (f *FakeAuthService) ChangePassword(_ context.Context, _ uint, _, _ string) error {
	return f.ChangePasswordErr
}
func (f *FakeAuthService) UpdateAvatar(_ context.Context, _ uint, _ string) error {
	return f.UpdateAvatarErr
}
func (f *FakeAuthService) Logout(_ context.Context, _ uint) error {
	return f.LogoutErr
}

// FakeUserService 单测用 IUserService mock
type FakeUserService struct {
	GetUsersList  []models.User
	GetUsersTotal int64
	GetUsersErr   error

	GetUserForAuthUser *models.User
	GetUserForAuthErr  error

	CreateUserResult *models.User
	CreateUserErr    error

	UpdateUserErr   error
	DeleteUserErr   error
	ResetPasswordPw string
	ResetPasswordErr error
	ToggleStatusErr error
}

func (f *FakeUserService) GetUsers(_ context.Context, _, _ int, _ map[string]string) ([]models.User, int64, error) {
	return f.GetUsersList, f.GetUsersTotal, f.GetUsersErr
}
func (f *FakeUserService) GetUserForAuth(_ context.Context, _ uint) (*models.User, error) {
	return f.GetUserForAuthUser, f.GetUserForAuthErr
}
func (f *FakeUserService) CreateUser(_ context.Context, _, _, _ string, _ int, _ string, _ []uint) (*models.User, error) {
	return f.CreateUserResult, f.CreateUserErr
}
func (f *FakeUserService) UpdateUser(_ context.Context, _ uint, _, _, _ string, _ []uint) error {
	return f.UpdateUserErr
}
func (f *FakeUserService) DeleteUser(_ context.Context, _ uint) error {
	return f.DeleteUserErr
}
func (f *FakeUserService) ResetPassword(_ context.Context, _ uint) (string, error) {
	return f.ResetPasswordPw, f.ResetPasswordErr
}
func (f *FakeUserService) ToggleStatus(_ context.Context, _ uint) error {
	return f.ToggleStatusErr
}

// FakeRoleService 单测用 IRoleService mock
type FakeRoleService struct {
	GetRolesList  []models.Role
	GetRolesTotal int64
	GetRolesErr   error

	CreateRoleResult *models.Role
	CreateRoleErr    error
	UpdateRoleResult *models.Role
	UpdateRoleErr    error
	DeleteRoleErr    error
	AssignPermissionsErr error
}

func (f *FakeRoleService) GetRoles(_ context.Context, _, _ int, _ map[string]string) ([]models.Role, int64, error) {
	return f.GetRolesList, f.GetRolesTotal, f.GetRolesErr
}
func (f *FakeRoleService) CreateRole(_ context.Context, _, _ string) (*models.Role, error) {
	return f.CreateRoleResult, f.CreateRoleErr
}
func (f *FakeRoleService) UpdateRole(_ context.Context, _ uint, _, _ string) (*models.Role, error) {
	return f.UpdateRoleResult, f.UpdateRoleErr
}
func (f *FakeRoleService) DeleteRole(_ context.Context, _ uint) error {
	return f.DeleteRoleErr
}
func (f *FakeRoleService) AssignPermissions(_ context.Context, _ uint, _ []uint) error {
	return f.AssignPermissionsErr
}

// FakePermissionService 单测用 IPermissionService mock
type FakePermissionService struct {
	GetPermissionsList  []models.Permission
	GetPermissionsTotal int64
	GetPermissionsErr   error

	CreatePermissionResult *models.Permission
	CreatePermissionErr    error
	UpdatePermissionResult *models.Permission
	UpdatePermissionErr    error
	DeletePermissionErr    error
	BatchDeletePermissionsErr error
	GetPermissionsByRoleIDsList []models.Permission
	GetPermissionsByRoleIDsErr  error
}

func (f *FakePermissionService) GetPermissions(_ context.Context, _, _ int, _ map[string]string) ([]models.Permission, int64, error) {
	return f.GetPermissionsList, f.GetPermissionsTotal, f.GetPermissionsErr
}
func (f *FakePermissionService) CreatePermission(_ context.Context, _, _, _, _, _ string) (*models.Permission, error) {
	return f.CreatePermissionResult, f.CreatePermissionErr
}
func (f *FakePermissionService) UpdatePermission(_ context.Context, _ uint, _, _, _ string) (*models.Permission, error) {
	return f.UpdatePermissionResult, f.UpdatePermissionErr
}
func (f *FakePermissionService) DeletePermission(_ context.Context, _ uint) error {
	return f.DeletePermissionErr
}
func (f *FakePermissionService) BatchDeletePermissions(_ context.Context, _ []uint) error {
	return f.BatchDeletePermissionsErr
}
func (f *FakePermissionService) GetPermissionsByRoleIDs(_ context.Context, _ []uint) ([]models.Permission, error) {
	return f.GetPermissionsByRoleIDsList, f.GetPermissionsByRoleIDsErr
}

// FakeOperationLogService 单测用 IOperationLogService mock
type FakeOperationLogService struct {
	GetOperationLogsList  []models.OperationLog
	GetOperationLogsTotal int64
	GetOperationLogsErr   error
	CleanOldLogsN        int64
	CleanOldLogsErr      error
}

func (f *FakeOperationLogService) GetOperationLogs(_ context.Context, _, _ int, _ map[string]string) ([]models.OperationLog, int64, error) {
	return f.GetOperationLogsList, f.GetOperationLogsTotal, f.GetOperationLogsErr
}
func (f *FakeOperationLogService) CleanOldLogs(_ context.Context, _ int) (int64, error) {
	return f.CleanOldLogsN, f.CleanOldLogsErr
}

// FakeDictionaryService 单测用 IDictionaryService mock
type FakeDictionaryService struct {
	GetTypesList  []models.DictType
	GetTypesTotal int64
	GetTypesErr   error

	CreateTypeResult *models.DictType
	CreateTypeErr    error
	UpdateTypeResult *models.DictType
	UpdateTypeErr    error
	DeleteTypeErr    error

	GetItemsList  []models.DictItem
	GetItemsTotal int64
	GetItemsErr   error
	GetItemsByCodeList []models.DictItem
	GetItemsByCodeErr  error
	CreateItemResult *models.DictItem
	CreateItemErr    error
	UpdateItemResult *models.DictItem
	UpdateItemErr    error
	DeleteItemErr    error
}

func (f *FakeDictionaryService) GetTypes(_ context.Context, _, _ int, _ map[string]string) ([]models.DictType, int64, error) {
	return f.GetTypesList, f.GetTypesTotal, f.GetTypesErr
}
func (f *FakeDictionaryService) CreateType(_ context.Context, _, _, _ string) (*models.DictType, error) {
	return f.CreateTypeResult, f.CreateTypeErr
}
func (f *FakeDictionaryService) UpdateType(_ context.Context, _ uint, _, _, _ string) (*models.DictType, error) {
	return f.UpdateTypeResult, f.UpdateTypeErr
}
func (f *FakeDictionaryService) DeleteType(_ context.Context, _ uint) error {
	return f.DeleteTypeErr
}
func (f *FakeDictionaryService) GetItems(_ context.Context, _ uint, _ string, _, _ int, _ map[string]string) ([]models.DictItem, int64, error) {
	return f.GetItemsList, f.GetItemsTotal, f.GetItemsErr
}
func (f *FakeDictionaryService) GetItemsByCode(_ context.Context, _ string) ([]models.DictItem, error) {
	return f.GetItemsByCodeList, f.GetItemsByCodeErr
}
func (f *FakeDictionaryService) CreateItem(_ context.Context, _ uint, _, _ string, _ int, _ int, _ string) (*models.DictItem, error) {
	return f.CreateItemResult, f.CreateItemErr
}
func (f *FakeDictionaryService) UpdateItem(_ context.Context, _ uint, _, _ string, _ *int, _ *int, _ string) (*models.DictItem, error) {
	return f.UpdateItemResult, f.UpdateItemErr
}
func (f *FakeDictionaryService) DeleteItem(_ context.Context, _ uint) error {
	return f.DeleteItemErr
}
