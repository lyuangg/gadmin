package testutil

import (
	"testing"

	"github.com/lyuangg/gadmin/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// NewTestDB 创建内存 SQLite DB 并执行迁移，不写入默认数据。供 services、controllers 等单测共用。
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{TablePrefix: ""},
		Logger:                                   gormlogger.Discard,
	})
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
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
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}
