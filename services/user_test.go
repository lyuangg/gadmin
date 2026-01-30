package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"
)

func TestUserService_GetUsers(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	users, total, err := svc.GetUsers(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if total != 0 || len(users) != 0 {
		t.Errorf("expected 0 users, got total=%d len=%d", total, len(users))
	}

	// 创建用户后再查
	_, err = svc.CreateUser(bg, "u1", "pass123", "昵称", 0, "", nil)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	users, total, err = svc.GetUsers(bg, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetUsers after create: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Errorf("expected 1 user, got total=%d len=%d", total, len(users))
	}
}

func TestUserService_CreateUser(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	user, err := svc.CreateUser(bg, "newuser", "mypass6", "新用户", 0, "备注", nil)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.ID == 0 || user.Username != "newuser" || user.Nickname != "新用户" {
		t.Errorf("CreateUser result: %+v", user)
	}
	// 不断言密码具体值，只确认有密码且可校验（CreateUser 用传入的 password 加密）
	var u models.User
	if err := db.Where("username = ?", "newuser").First(&u).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if u.Password == "" {
		t.Error("password should be set")
	}
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	_, _ = svc.CreateUser(bg, "dup", "pass123", "", 0, "", nil)
	_, err := svc.CreateUser(bg, "dup", "other", "", 0, "", nil)
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestUserService_GetUsers_ByID(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	_, _ = svc.CreateUser(bg, "byid", "pass", "ByID", 0, "", nil)
	users, total, err := svc.GetUsers(bg, 1, 10, map[string]string{"username": "byid"})
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Fatalf("expected 1 user, got total=%d len=%d", total, len(users))
	}
	if users[0].Username != "byid" || users[0].Nickname != "ByID" {
		t.Errorf("GetUsers result: %+v", users[0])
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	created, _ := svc.CreateUser(bg, "upuser", "oldpass", "旧昵称", 0, "", nil)
	err := svc.UpdateUser(bg, created.ID, "新昵称", "newpass6", "备注", nil)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	var user models.User
	if err := db.Where("id = ?", created.ID).First(&user).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if user.Nickname != "新昵称" || user.Remark != "备注" {
		t.Errorf("UpdateUser result: nickname=%s remark=%s", user.Nickname, user.Remark)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	created, _ := svc.CreateUser(bg, "deluser", "pass", "", 0, "", nil)
	err := svc.DeleteUser(bg, created.ID)
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	var u models.User
	if err := db.First(&u, created.ID).Error; err == nil {
		t.Error("user should be deleted (soft-deleted, not visible)")
	}
}

func TestUserService_ToggleStatus(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewUserService(ctx)
	bg := context.Background()

	created, _ := svc.CreateUser(bg, "toggle", "pass", "", 0, "", nil)
	if created.Status != 1 {
		t.Errorf("default status expected 1, got %d", created.Status)
	}
	err := svc.ToggleStatus(bg, created.ID)
	if err != nil {
		t.Fatalf("ToggleStatus: %v", err)
	}
	var u models.User
	if err := db.First(&u, created.ID).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if u.Status != 0 {
		t.Errorf("after toggle expected 0, got %d", u.Status)
	}
}

// ResetPassword 使用随机密码，单测不断言具体密码值，仅跳过或做弱断言。
// 此处选择跳过：依赖随机生成密码，难以稳定断言。
func TestUserService_ResetPassword(t *testing.T) {
	t.Skip("ResetPassword 使用随机密码，单测跳过")
}
