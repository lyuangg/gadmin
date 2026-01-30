package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"
)

func TestPermissionService_GetPermissions(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	list, total, err := svc.GetPermissions(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetPermissions: %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Errorf("expected 0 permissions, got total=%d len=%d", total, len(list))
	}

	_, _ = svc.CreatePermission(bg, "/api/users", "GET", "用户列表", "用户管理", "")
	list, total, err = svc.GetPermissions(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetPermissions: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 permission, got total=%d len=%d", total, len(list))
	}
}

func TestPermissionService_CreatePermission(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	p, err := svc.CreatePermission(bg, "/api/roles", "POST", "创建角色", "角色管理", "描述")
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	if p.ID == 0 || p.Path != "/api/roles" || p.Method != "POST" || p.Name != "创建角色" {
		t.Errorf("CreatePermission result: %+v", p)
	}

	_, err = svc.CreatePermission(bg, "/api/roles", "POST", "重复", "", "")
	if err == nil {
		t.Error("expected error for duplicate path+method")
	}
}

func TestPermissionService_UpdatePermission(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	created, _ := svc.CreatePermission(bg, "/update", "PUT", "旧名", "旧组", "")
	updated, err := svc.UpdatePermission(bg, created.ID, "新名", "新组", "新描述")
	if err != nil {
		t.Fatalf("UpdatePermission: %v", err)
	}
	if updated.Name != "新名" || updated.Group != "新组" || updated.Description != "新描述" {
		t.Errorf("UpdatePermission result: %+v", updated)
	}

	_, err = svc.UpdatePermission(bg, 99999, "x", "", "")
	if err == nil {
		t.Error("expected error for non-existent permission")
	}
}

func TestPermissionService_DeletePermission(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	created, _ := svc.CreatePermission(bg, "/del", "DELETE", "待删", "", "")
	err := svc.DeletePermission(bg, created.ID)
	if err != nil {
		t.Fatalf("DeletePermission: %v", err)
	}
	var p models.Permission
	if err := db.First(&p, created.ID).Error; err == nil {
		t.Error("permission should be deleted")
	}
}

func TestPermissionService_GetPermissionsByRoleIDs(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	// 空角色列表
	perms, err := svc.GetPermissionsByRoleIDs(bg, nil)
	if err != nil {
		t.Fatalf("GetPermissionsByRoleIDs nil: %v", err)
	}
	if perms != nil {
		t.Errorf("expected nil, got len=%d", len(perms))
	}

	// 创建角色和权限并关联
	role, _ := NewRoleService(ctx).CreateRole(bg, "r1", "")
	p1, _ := svc.CreatePermission(bg, "/p1", "GET", "P1", "", "")
	_ = NewRoleService(ctx).AssignPermissions(bg, role.ID, []uint{p1.ID})

	perms, err = svc.GetPermissionsByRoleIDs(bg, []uint{role.ID})
	if err != nil {
		t.Fatalf("GetPermissionsByRoleIDs: %v", err)
	}
	if len(perms) != 1 || perms[0].ID != p1.ID {
		t.Errorf("expected 1 permission, got %+v", perms)
	}
}

func TestPermissionService_BatchDeletePermissions(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewPermissionService(ctx)
	bg := context.Background()

	p1, _ := svc.CreatePermission(bg, "/b1", "GET", "B1", "", "")
	p2, _ := svc.CreatePermission(bg, "/b2", "GET", "B2", "", "")

	err := svc.BatchDeletePermissions(bg, []uint{})
	if err == nil {
		t.Error("expected error for empty ids")
	}

	err = svc.BatchDeletePermissions(bg, []uint{p1.ID, p2.ID})
	if err != nil {
		t.Fatalf("BatchDeletePermissions: %v", err)
	}
	var count int64
	db.Model(&models.Permission{}).Where("id IN ?", []uint{p1.ID, p2.ID}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 remaining, got %d", count)
	}
}
