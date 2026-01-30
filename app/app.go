package app

import (
	"io"

	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/database"
	"github.com/lyuangg/gadmin/logger"
	"github.com/lyuangg/gadmin/response"
	"github.com/lyuangg/gadmin/services"

	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	db     *gorm.DB
	logger logger.ILogger

	// Close() 会依次调用已注册的 Closer
	closers []io.Closer

	Responder response.IResponder

	CaptchaProvider services.CaptchaProvider
	TokenGenerator services.TokenGenerator

	AuthService         services.IAuthService
	UserService         services.IUserService
	RoleService         services.IRoleService
	PermissionService   services.IPermissionService
	OperationLogService services.IOperationLogService
	DictionaryService   services.IDictionaryService
}

// NewApp 若初始化失败会 panic
func NewApp(cfg *config.Config) *App {
	slogLogger, logHandler := logger.NewLogger(cfg)

	db, err := database.InitDB(cfg, slogLogger)
	if err != nil {
		panic("数据库初始化失败: " + err.Error())
	}

	app := &App{
		Config:    cfg,
		db:        db,
		logger:    slogLogger,
		closers:   []io.Closer{logHandler},
		Responder: response.NewResponder(slogLogger),
	}

	app.CaptchaProvider = services.NewRealCaptchaProvider()
	app.TokenGenerator = services.NewRealTokenGenerator()

	app.AuthService = services.NewAuthService(app)
	app.UserService = services.NewUserService(app)
	app.RoleService = services.NewRoleService(app)
	app.PermissionService = services.NewPermissionService(app)
	app.OperationLogService = services.NewOperationLogService(app)
	app.DictionaryService = services.NewDictionaryService(app)

	return app
}

func (a *App) DB() *gorm.DB {
	return a.db
}

func (a *App) Logger() logger.ILogger {
	return a.logger
}

func (a *App) GetConfig() *config.Config {
	return a.Config
}

func (a *App) GetCaptchaProvider() services.CaptchaProvider {
	return a.CaptchaProvider
}

func (a *App) GetTokenGenerator() services.TokenGenerator {
	return a.TokenGenerator
}

func (a *App) GetAuthService() services.IAuthService {
	return a.AuthService
}

func (a *App) GetUserService() services.IUserService {
	return a.UserService
}

func (a *App) GetRoleService() services.IRoleService {
	return a.RoleService
}

func (a *App) GetPermissionService() services.IPermissionService {
	return a.PermissionService
}

func (a *App) GetOperationLogService() services.IOperationLogService {
	return a.OperationLogService
}

func (a *App) GetDictionaryService() services.IDictionaryService {
	return a.DictionaryService
}

// RegisterCloser 注册退出时需关闭的对象
func (a *App) RegisterCloser(c io.Closer) {
	if c != nil {
		a.closers = append(a.closers, c)
	}
}

// Close 依次调用已注册的 Closer，应在程序退出前调用
func (a *App) Close() error {
	var firstErr error
	for _, c := range a.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
