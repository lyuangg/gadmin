package services

import (
	"log/slog"
	"testing"

	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/internal/testutil"
	"github.com/lyuangg/gadmin/logger"

	"gorm.io/gorm"
)

// testServiceContext 单测用 ServiceContext，仅提供 DB / Captcha / TokenGenerator，其他 Service 返回 nil
type testServiceContext struct {
	db       *gorm.DB
	captcha  CaptchaProvider
	tokenGen TokenGenerator
	logger   logger.ILogger
	cfg      *config.Config
	auth     *AuthService
	user     *UserService
	role     *RoleService
	perm     *PermissionService
	opLog    *OperationLogService
}

func (c *testServiceContext) DB() *gorm.DB                                 { return c.db }
func (c *testServiceContext) Logger() logger.ILogger                       { return c.logger }
func (c *testServiceContext) GetConfig() *config.Config                    { return c.cfg }
func (c *testServiceContext) GetCaptchaProvider() CaptchaProvider          { return c.captcha }
func (c *testServiceContext) GetTokenGenerator() TokenGenerator            { return c.tokenGen }
func (c *testServiceContext) GetAuthService() IAuthService                 { return c.auth }
func (c *testServiceContext) GetUserService() IUserService                 { return c.user }
func (c *testServiceContext) GetRoleService() IRoleService                 { return c.role }
func (c *testServiceContext) GetPermissionService() IPermissionService     { return c.perm }
func (c *testServiceContext) GetOperationLogService() IOperationLogService { return c.opLog }

// NewTestDB 委托给 testutil，保持 services 包内单测调用不变
func NewTestDB(t *testing.T) *gorm.DB {
	return testutil.NewTestDB(t)
}

// NewTestServiceContext 创建单测用 ServiceContext：DB 为内存库，Captcha/Token 为 Fake
func NewTestServiceContext(t *testing.T, db *gorm.DB, opts ...TestContextOption) ServiceContext {
	t.Helper()
	ctx := &testServiceContext{
		db:       db,
		logger:   slog.Default(),
		cfg:      &config.Config{},
		captcha:  &FakeCaptchaProvider{VerifyResult: true},
		tokenGen: &FakeTokenGenerator{Token: "fake-token"},
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// TestContextOption 可选配置 testServiceContext
type TestContextOption func(*testServiceContext)

// WithCaptchaProvider 指定验证码提供者
func WithCaptchaProvider(p CaptchaProvider) TestContextOption {
	return func(c *testServiceContext) { c.captcha = p }
}

// WithTokenGenerator 指定 Token 生成器
func WithTokenGenerator(tg TokenGenerator) TestContextOption {
	return func(c *testServiceContext) { c.tokenGen = tg }
}

// WithConfig 指定配置
func WithConfig(cfg *config.Config) TestContextOption {
	return func(c *testServiceContext) { c.cfg = cfg }
}
