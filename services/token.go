package services

import (
	"github.com/lyuangg/gadmin/utils"
)

// TokenGenerator JWT Token 生成器，便于测试时替换为 mock
type TokenGenerator interface {
	GenerateToken(userID uint, username, nickname string, userType int, isSuperAdmin bool, roleIDs []uint, tokenVersion uint) (string, error)
}

// realTokenGenerator 生产实现，委托 utils.GenerateToken（需在 main 中先调用 utils.InitJWT）
type realTokenGenerator struct{}

// NewRealTokenGenerator 创建生产用 Token 生成器
func NewRealTokenGenerator() TokenGenerator {
	return &realTokenGenerator{}
}

func (t *realTokenGenerator) GenerateToken(userID uint, username, nickname string, userType int, isSuperAdmin bool, roleIDs []uint, tokenVersion uint) (string, error) {
	return utils.GenerateToken(userID, username, nickname, userType, isSuperAdmin, roleIDs, tokenVersion)
}

// FakeTokenGenerator 单测用，返回固定 token 或错误
type FakeTokenGenerator struct {
	Token string
	Err   error
}

func (f *FakeTokenGenerator) GenerateToken(userID uint, username, nickname string, userType int, isSuperAdmin bool, roleIDs []uint, tokenVersion uint) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	return f.Token, nil
}
