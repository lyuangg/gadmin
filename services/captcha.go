package services

import (
	"context"

	"github.com/mojocn/base64Captcha"
)

// CaptchaProvider 验证码提供者，便于测试时替换为 mock
type CaptchaProvider interface {
	Generate(ctx context.Context) (id, b64s string, err error)
	Verify(id, answer string) bool
}

// realCaptcha 生产环境实现，使用 base64Captcha
type realCaptcha struct {
	store base64Captcha.Store
}

// NewRealCaptchaProvider 创建生产用验证码提供者
func NewRealCaptchaProvider() CaptchaProvider {
	return &realCaptcha{store: base64Captcha.DefaultMemStore}
}

func (c *realCaptcha) Generate(ctx context.Context) (id, b64s string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, c.store)
	id, b64s, _, err = captcha.Generate()
	return id, b64s, err
}

func (c *realCaptcha) Verify(id, answer string) bool {
	return c.store.Verify(id, answer, true)
}

// FakeCaptchaProvider 单测用，Verify/Generate 行为可配置
type FakeCaptchaProvider struct {
	VerifyResult bool
	GenerateID   string
	GenerateB64  string
	GenerateErr  error
}

func (f *FakeCaptchaProvider) Generate(ctx context.Context) (id, b64s string, err error) {
	return f.GenerateID, f.GenerateB64, f.GenerateErr
}

func (f *FakeCaptchaProvider) Verify(id, answer string) bool {
	return f.VerifyResult
}
