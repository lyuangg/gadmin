package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"
)

func TestOperationLogService_GetOperationLogs(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewOperationLogService(ctx)
	bg := context.Background()

	logs, total, err := svc.GetOperationLogs(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetOperationLogs: %v", err)
	}
	if total != 0 || len(logs) != 0 {
		t.Errorf("expected 0 logs, got total=%d len=%d", total, len(logs))
	}

	// 插入一条日志
	if err := db.Create(&models.OperationLog{
		UserID:     1,
		Username:   "test",
		Method:     "GET",
		Path:       "/api/test",
		StatusCode: 200,
	}).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}
	logs, total, err = svc.GetOperationLogs(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetOperationLogs: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Errorf("expected 1 log, got total=%d len=%d", total, len(logs))
	}
	if logs[0].Method != "GET" || logs[0].Path != "/api/test" {
		t.Errorf("log mismatch: %+v", logs[0])
	}
}

func TestOperationLogService_GetOperationLogs_Filter(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewOperationLogService(ctx)
	bg := context.Background()

	_ = db.Create(&models.OperationLog{UserID: 0, Username: "a", Method: "GET", Path: "/a", StatusCode: 200}).Error
	_ = db.Create(&models.OperationLog{UserID: 0, Username: "b", Method: "POST", Path: "/b", StatusCode: 201}).Error

	logs, total, err := svc.GetOperationLogs(bg, 1, 10, map[string]string{"method": "POST"})
	if err != nil {
		t.Fatalf("GetOperationLogs filter: %v", err)
	}
	if total != 1 || len(logs) != 1 || logs[0].Method != "POST" {
		t.Errorf("filter by method: total=%d len=%d", total, len(logs))
	}
}

func TestOperationLogService_CleanOldLogs(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewOperationLogService(ctx)
	bg := context.Background()

	// retain=0 不删
	n, err := svc.CleanOldLogs(bg, 0)
	if err != nil {
		t.Fatalf("CleanOldLogs(0): %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}

	// 插入 5 条
	for i := 0; i < 5; i++ {
		_ = db.Create(&models.OperationLog{UserID: 0, Username: "u", Method: "GET", Path: "/", StatusCode: 200}).Error
	}
	var count int64
	db.Model(&models.OperationLog{}).Count(&count)
	if count != 5 {
		t.Fatalf("expected 5 logs, got %d", count)
	}

	// 保留最近 2 条，应删 3 条
	n, err = svc.CleanOldLogs(bg, 2)
	if err != nil {
		t.Fatalf("CleanOldLogs(2): %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 deleted, got %d", n)
	}
	db.Model(&models.OperationLog{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 remaining, got %d", count)
	}
}
