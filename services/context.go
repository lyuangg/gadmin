package services

import (
	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/logger"

	"gorm.io/gorm"
)

type ServiceContext interface {
	DB() *gorm.DB
	Logger() logger.ILogger
	GetConfig() *config.Config
	GetCaptchaProvider() CaptchaProvider
	GetTokenGenerator() TokenGenerator
	GetAuthService() IAuthService
	GetUserService() IUserService
	GetRoleService() IRoleService
	GetPermissionService() IPermissionService
	GetOperationLogService() IOperationLogService
}
