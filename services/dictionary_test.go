package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"
)

func TestDictionaryService_GetTypes(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	list, total, err := svc.GetTypes(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetTypes: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Errorf("expected 0 types, got total=%d len=%d", total, len(list))
	}

	_, _ = svc.CreateType(bg, "status", "状态", "")
	list, total, err = svc.GetTypes(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetTypes after create: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 type, got total=%d len=%d", total, len(list))
	}
}

func TestDictionaryService_CreateType(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	dt, err := svc.CreateType(bg, "gender", "性别", "备注")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}
	if dt.ID == 0 || dt.Code != "gender" || dt.Name != "性别" {
		t.Errorf("CreateType result: %+v", dt)
	}

	_, err = svc.CreateType(bg, "gender", "其他", "")
	if err == nil {
		t.Error("expected error for duplicate code")
	}
}

func TestDictionaryService_UpdateType(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	created, _ := svc.CreateType(bg, "old", "旧名", "")
	updated, err := svc.UpdateType(bg, created.ID, "new", "新名", "备注")
	if err != nil {
		t.Fatalf("UpdateType: %v", err)
	}
	if updated.Code != "new" || updated.Name != "新名" {
		t.Errorf("UpdateType result: %+v", updated)
	}

	_, err = svc.UpdateType(bg, 99999, "x", "", "")
	if err == nil {
		t.Error("expected error for non-existent type")
	}
}

func TestDictionaryService_DeleteType(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	created, _ := svc.CreateType(bg, "del", "待删", "")
	err := svc.DeleteType(bg, created.ID)
	if err != nil {
		t.Fatalf("DeleteType: %v", err)
	}
	var dt models.DictType
	if err := db.First(&dt, created.ID).Error; err == nil {
		t.Error("type should be deleted")
	}
}

func TestDictionaryService_GetItems_CreateItem_UpdateItem_DeleteItem(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	dt, _ := svc.CreateType(bg, "test_items", "测试项", "")

	// GetItems 空
	items, total, err := svc.GetItems(bg, dt.ID, "", 1, 10, nil)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Errorf("expected 0 items, got total=%d len=%d", total, len(items))
	}

	// CreateItem
	item, err := svc.CreateItem(bg, dt.ID, "男", "1", 0, 1, "")
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if item.ID == 0 || item.Label != "男" || item.Value != "1" {
		t.Errorf("CreateItem result: %+v", item)
	}

	// 同类型下 value 唯一
	_, err = svc.CreateItem(bg, dt.ID, "男2", "1", 1, 1, "")
	if err == nil {
		t.Error("expected error for duplicate value in same type")
	}

	// GetItems 有数据
	items, total, err = svc.GetItems(bg, dt.ID, "", 1, 10, nil)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got total=%d len=%d", total, len(items))
	}

	// GetItemsByCode
	itemsByCode, err := svc.GetItemsByCode(bg, "test_items")
	if err != nil {
		t.Fatalf("GetItemsByCode: %v", err)
	}
	if len(itemsByCode) != 1 || itemsByCode[0].Value != "1" {
		t.Errorf("GetItemsByCode result: %+v", itemsByCode)
	}

	// UpdateItem
	sortVal := 10
	statusVal := 0
	updated, err := svc.UpdateItem(bg, item.ID, "男性", "1", &sortVal, &statusVal, "备注")
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if updated.Label != "男性" || updated.Sort != 10 || updated.Status != 0 {
		t.Errorf("UpdateItem result: %+v", updated)
	}

	// DeleteItem
	err = svc.DeleteItem(bg, item.ID)
	if err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	items, _, _ = svc.GetItems(bg, dt.ID, "", 1, 10, nil)
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestDictionaryService_GetItems_TypeCode(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewDictionaryService(ctx)
	bg := context.Background()

	_, _ = svc.CreateType(bg, "bycode", "按编码", "")
	items, _, err := svc.GetItems(bg, 0, "bycode", 1, 10, nil)
	if err != nil {
		t.Fatalf("GetItems by type_code: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}

	_, _, err = svc.GetItems(bg, 0, "", 1, 10, nil)
	if err == nil {
		t.Error("expected error when neither type_id nor type_code")
	}
}
