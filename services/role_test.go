package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"
)

func TestRoleService_GetRoles(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewRoleService(ctx)
	bg := context.Background()

	// 空表
	roles, total, err := svc.GetRoles(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetRoles empty: %v", err)
	}
	if total != 0 || len(roles) != 0 {
		t.Errorf("expected 0 roles, got total=%d len=%d", total, len(roles))
	}

	// 创建一条再查
	created, err := svc.CreateRole(bg, "测试角色", "描述")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	roles, total, err = svc.GetRoles(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetRoles after create: %v", err)
	}
	if total != 1 || len(roles) != 1 {
		t.Errorf("expected 1 role, got total=%d len=%d", total, len(roles))
	}
	if roles[0].Name != "测试角色" || roles[0].ID != created.ID {
		t.Errorf("role mismatch: got %+v", roles[0])
	}
}

func TestRoleService_CreateRole(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewRoleService(ctx)
	bg := context.Background()

	role, err := svc.CreateRole(bg, "管理员", "系统管理员")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.ID == 0 || role.Name != "管理员" || role.Description != "系统管理员" {
		t.Errorf("CreateRole result: %+v", role)
	}

	// 同名应失败
	_, err = svc.CreateRole(bg, "管理员", "")
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestRoleService_UpdateRole(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewRoleService(ctx)
	bg := context.Background()

	created, _ := svc.CreateRole(bg, "旧名", "旧描述")
	updated, err := svc.UpdateRole(bg, created.ID, "新名", "新描述")
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if updated.Name != "新名" || updated.Description != "新描述" {
		t.Errorf("UpdateRole result: %+v", updated)
	}

	// 不存在的 ID
	_, err = svc.UpdateRole(bg, 99999, "x", "")
	if err == nil {
		t.Error("expected error for non-existent role")
	}
}

func TestRoleService_DeleteRole(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewRoleService(ctx)
	bg := context.Background()

	created, _ := svc.CreateRole(bg, "待删", "")
	err := svc.DeleteRole(bg, created.ID)
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	var r models.Role
	if err := db.First(&r, created.ID).Error; err == nil {
		t.Error("role should be deleted")
	}
}

func TestRoleService_AssignPermissions(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewRoleService(ctx)
	bg := context.Background()

	// 先建一个权限
	var perm models.Permission
	perm.Path = "/api/test"
	perm.Method = "GET"
	perm.Name = "测试权限"
	if err := db.Create(&perm).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}

	role, _ := svc.CreateRole(bg, "有权限角色", "")
	err := svc.AssignPermissions(bg, role.ID, []uint{perm.ID})
	if err != nil {
		t.Fatalf("AssignPermissions: %v", err)
	}
	count := db.Model(&models.Role{ID: role.ID}).Association("Permissions").Count()
	if count != 1 {
		t.Errorf("expected 1 permission, got %d", count)
	}
}
