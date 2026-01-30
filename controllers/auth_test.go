package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
)

func TestAuthController_Login_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{
		AuthService: &services.FakeAuthService{},
	})
	ctrl := NewAuthController(a)

	body := []byte(`{}`) // 缺少 required 字段
	c, w := newGinContext(http.MethodPost, "/api/login", body)
	ctrl.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error response (code != 0)")
	}
}

func TestAuthController_Login_Success(t *testing.T) {
	authMock := &services.FakeAuthService{
		LoginUser:  &models.User{ID: 1, Username: "ctrluser", Nickname: "测试", Avatar: "", Type: 0, Status: 1, Roles: nil},
		LoginToken: "test-token",
		LoginErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{AuthService: authMock})
	ctrl := NewAuthController(a)

	body, _ := json.Marshal(map[string]string{
		"username":    "ctrluser",
		"password":    "pass123",
		"captcha_id":  "any",
		"captcha_val": "any",
	})
	c, w := newGinContext(http.MethodPost, "/api/login", body)
	ctrl.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v body=%s", resp["code"], w.Body.Bytes())
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		t.Fatal("expected data")
	}
	if data["token"] != "test-token" {
		t.Errorf("expected token test-token, got %v", data["token"])
	}
}

func TestAuthController_GetCaptcha(t *testing.T) {
	authMock := &services.FakeAuthService{
		GenerateID:  "cid",
		GenerateB64: "data:image/png;base64,xx",
		GenerateErr: nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{AuthService: authMock})
	ctrl := NewAuthController(a)

	c, w := newGinContextGET("/api/captcha")
	ctrl.GetCaptcha(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
	data, _ := resp["data"].(map[string]interface{})
	if data["captcha_id"] != "cid" || data["captcha_img"] != "data:image/png;base64,xx" {
		t.Errorf("captcha mismatch: %v", data)
	}
}

func TestAuthController_Logout_Unauthorized(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{AuthService: &services.FakeAuthService{}})
	ctrl := NewAuthController(a)

	c, w := newGinContext(http.MethodPost, "/api/logout", nil) // 未 Set("user")
	ctrl.Logout(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error when no user in context")
	}
}

func TestAuthController_Logout_Success(t *testing.T) {
	authMock := &services.FakeAuthService{LogoutErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{AuthService: authMock})
	ctrl := NewAuthController(a)

	u := models.User{ID: 1, Username: "logoutu", Nickname: "登出", Type: 0, Status: 1}
	c, w := newGinContextWithUser(http.MethodPost, "/api/logout", nil, u)
	ctrl.Logout(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}
