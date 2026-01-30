package app

import (
	"log/slog"

	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/logger"
	"github.com/lyuangg/gadmin/response"
	"github.com/lyuangg/gadmin/services"

	"gorm.io/gorm"
)

// ServiceMocks 单测用：可注入的 service 接口 mock，仅需提供被测 controller 用到的 service
type ServiceMocks struct {
	AuthService         services.IAuthService
	UserService         services.IUserService
	RoleService         services.IRoleService
	PermissionService   services.IPermissionService
	OperationLogService services.IOperationLogService
	DictionaryService   services.IDictionaryService
}

// NewTestAppWithServiceMocks 供 controller 单测用：不设置 db，仅注入 mock service；未提供的 service 为 nil，调用会 panic。
func NewTestAppWithServiceMocks(mocks *ServiceMocks) *App {
	return newTestAppWithLoggerAndResponder(slog.Default(), response.NewResponder(slog.Default()), mocks)
}

// NewTestAppWithLogger 供中间件等单测用：注入 ILogger 以断言日志内容；Responder 使用默认实现。
func NewTestAppWithLogger(l logger.ILogger) *App {
	return newTestAppWithLoggerAndResponder(l, response.NewResponder(slog.Default()), nil)
}

func newTestAppWithLoggerAndResponder(l logger.ILogger, resp response.IResponder, mocks *ServiceMocks) *App {
	cfg := &config.Config{}
	a := &App{
		Config:          cfg,
		db:              nil,
		logger:          l,
		closers:         nil,
		Responder:       resp,
		CaptchaProvider: nil,
		TokenGenerator:  nil,
	}
	if mocks != nil {
		a.AuthService = mocks.AuthService
		a.UserService = mocks.UserService
		a.RoleService = mocks.RoleService
		a.PermissionService = mocks.PermissionService
		a.OperationLogService = mocks.OperationLogService
		a.DictionaryService = mocks.DictionaryService
	}
	return a
}

// NewTestApp 供测试用：使用已迁移的 db 构建 App（不连接真实数据库，不注册 Close）
func NewTestApp(db *gorm.DB) *App {
	return NewTestAppWithMocks(db, nil, nil)
}

// NewTestAppWithMocks 供测试用：可注入 CaptchaProvider、TokenGenerator（为 nil 时使用生产实现）
func NewTestAppWithMocks(db *gorm.DB, captcha services.CaptchaProvider, tokenGen services.TokenGenerator) *App {
	cfg := &config.Config{}
	slogLogger := slog.Default()
	if captcha == nil {
		captcha = services.NewRealCaptchaProvider()
	}
	if tokenGen == nil {
		tokenGen = services.NewRealTokenGenerator()
	}
	a := &App{
		Config:          cfg,
		db:              db,
		logger:          slogLogger,
		closers:         nil,
		Responder:       response.NewResponder(slogLogger),
		CaptchaProvider: captcha,
		TokenGenerator:  tokenGen,
	}
	a.AuthService = services.NewAuthService(a)
	a.UserService = services.NewUserService(a)
	a.RoleService = services.NewRoleService(a)
	a.PermissionService = services.NewPermissionService(a)
	a.OperationLogService = services.NewOperationLogService(a)
	a.DictionaryService = services.NewDictionaryService(a)
	return a
}
