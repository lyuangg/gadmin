package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type slogGormWriter struct {
	logger *slog.Logger
}

func (w *slogGormWriter) Printf(format string, args ...interface{}) {
	w.logger.InfoContext(context.Background(), fmt.Sprintf(format, args...))
}

func newGormLogger(cfg *config.Config, slogLogger *slog.Logger) gormlogger.Interface {
	level := gormlogger.Warn
	switch strings.ToLower(cfg.DBLogLevel) {
	case "silent":
		level = gormlogger.Silent
	case "error":
		level = gormlogger.Error
	case "warn", "warning":
		level = gormlogger.Warn
	case "info", "debug":
		level = gormlogger.Info
	}

	return gormlogger.New(
		&slogGormWriter{logger: slogLogger},
		gormlogger.Config{
			SlowThreshold:             time.Duration(cfg.DBSlowThresholdMs) * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

func InitDB(cfg *config.Config, slogLogger *slog.Logger) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{TablePrefix: cfg.DBTablePrefix},
		Logger:                                   newGormLogger(cfg, slogLogger),
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.OperationLog{},
		&models.DictType{},
		&models.DictItem{},
	)
	if err != nil {
		return nil, err
	}

	if err := initDefaultData(db, slogLogger); err != nil {
		return nil, err
	}

	return db, nil
}

func initDefaultData(db *gorm.DB, logger *slog.Logger) error {
	var superAdminRole models.Role
	result := db.Where("name = ?", "超级管理员").First(&superAdminRole)
	if result.Error == gorm.ErrRecordNotFound {
		superAdminRole = models.Role{
			Name:        "超级管理员",
			Description: "拥有所有权限的超级管理员",
		}
		if err := db.Create(&superAdminRole).Error; err != nil {
			return err
		}
		logger.InfoContext(context.Background(), "创建默认超级管理员角色")
	}

	var adminUser models.User
	result = db.Where("username = ?", "admin").First(&adminUser)
	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminUser = models.User{
			Username: "admin",
			Password: string(hashedPassword),
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}

		if err := db.Model(&adminUser).Association("Roles").Append(&superAdminRole); err != nil {
			return err
		}
		logger.InfoContext(context.Background(), "创建默认管理员账号", "username", "admin", "password", "admin123")
	}

	return nil
}
