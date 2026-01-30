package services

import (
	"context"
	"testing"

	"github.com/lyuangg/gadmin/models"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login_VerifyCaptchaFail(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db, WithCaptchaProvider(&FakeCaptchaProvider{VerifyResult: false}))
	svc := NewAuthService(ctx)
	bg := context.Background()

	_, _, err := svc.Login(bg, "admin", "admin123", "cid", "val")
	if err == nil {
		t.Error("expected error when captcha fails")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	db := NewTestDB(t)
	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	_, _, err := svc.Login(bg, "nonexistent", "any", "cid", "val")
	if err == nil {
		t.Error("expected error when user not found")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	db := NewTestDB(t)
	// 插入一个已知密码的用户
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	user := models.User{
		Username:     "testuser",
		Password:     string(hashed),
		Nickname:     "测试",
		Type:         0,
		Status:       1,
		TokenVersion: 0,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := NewTestServiceContext(t, db, WithTokenGenerator(&FakeTokenGenerator{Token: "my-token"}))
	svc := NewAuthService(ctx)
	bg := context.Background()

	u, token, err := svc.Login(bg, "testuser", "pass123", "cid", "val")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.Username != "testuser" || token != "my-token" {
		t.Errorf("user=%s token=%s", u.Username, token)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	db := NewTestDB(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)
	if err := db.Create(&models.User{
		Username: "u2",
		Password: string(hashed),
		Type:     0,
		Status:   1,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	_, _, err := svc.Login(bg, "u2", "wrong", "cid", "val")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	db := NewTestDB(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	var user models.User
	user.Username = "cpuser"
	user.Password = string(hashed)
	user.Type = 0
	user.Status = 1
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	err := svc.ChangePassword(bg, user.ID, "oldpass", "newpass6")
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	var u models.User
	if err := db.First(&u, user.ID).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newpass6")) != nil {
		t.Error("password was not updated correctly")
	}
}

func TestAuthService_ChangePassword_WrongOld(t *testing.T) {
	db := NewTestDB(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.DefaultCost)
	if err := db.Create(&models.User{Username: "x", Password: string(hashed), Type: 0, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	var u models.User
	db.Where("username = ?", "x").First(&u)

	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	err := svc.ChangePassword(bg, u.ID, "wrongold", "newpass")
	if err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestAuthService_UpdateAvatar(t *testing.T) {
	db := NewTestDB(t)
	if err := db.Create(&models.User{Username: "av", Password: "x", Type: 0, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	var u models.User
	db.Where("username = ?", "av").First(&u)

	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	err := svc.UpdateAvatar(bg, u.ID, "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("UpdateAvatar: %v", err)
	}
	db.First(&u, u.ID)
	if u.Avatar != "https://example.com/avatar.png" {
		t.Errorf("avatar not updated: %s", u.Avatar)
	}
}

func TestAuthService_GenerateCaptcha(t *testing.T) {
	ctx := NewTestServiceContext(t, NewTestDB(t), WithCaptchaProvider(&FakeCaptchaProvider{
		GenerateID: "fake-id", GenerateB64: "data:image/png;base64,xxx",
	}))
	svc := NewAuthService(ctx)
	bg := context.Background()

	id, b64, err := svc.GenerateCaptcha(bg)
	if err != nil {
		t.Fatalf("GenerateCaptcha: %v", err)
	}
	if id != "fake-id" || b64 != "data:image/png;base64,xxx" {
		t.Errorf("id=%s b64=%s", id, b64)
	}
}

func TestAuthService_Logout(t *testing.T) {
	db := NewTestDB(t)
	if err := db.Create(&models.User{Username: "logoutuser", Password: "x", Type: 0, Status: 1, TokenVersion: 0}).Error; err != nil {
		t.Fatal(err)
	}
	var u models.User
	db.Where("username = ?", "logoutuser").First(&u)

	ctx := NewTestServiceContext(t, db)
	svc := NewAuthService(ctx)
	bg := context.Background()

	err := svc.Logout(bg, u.ID)
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}
	db.First(&u, u.ID)
	if u.TokenVersion != 1 {
		t.Errorf("expected TokenVersion 1, got %d", u.TokenVersion)
	}
}
