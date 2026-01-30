package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
)

func TestUserController_GetUsers(t *testing.T) {
	userMock := &services.FakeUserService{
		GetUsersList:  []models.User{},
		GetUsersTotal: 0,
		GetUsersErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	c, w := newGinContextGET("/api/users?page=1&page_size=10")
	ctrl.GetUsers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		t.Fatal("expected data")
	}
	if _, ok := data["data"]; !ok {
		t.Error("expected data.data (list)")
	}
	if _, ok := data["pagination"]; !ok {
		t.Error("expected data.pagination")
	}
}

func TestUserController_CreateUser(t *testing.T) {
	userMock := &services.FakeUserService{
		CreateUserResult: &models.User{ID: 1, Username: "testuser", Nickname: "测试用户", Type: 0, Status: 1},
		CreateUserErr:    nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	body, _ := json.Marshal(map[string]interface{}{
		"username": "testuser",
		"password": "pass123",
		"nickname": "测试用户",
		"type":     0,
		"remark":   "备注",
		"role_ids": []uint{},
	})
	c, w := newGinContext(http.MethodPost, "/api/users", body)
	ctrl.CreateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		t.Fatal("expected data")
	}
	if data["username"] != "testuser" {
		t.Errorf("create result: %v", data)
	}
}

func TestUserController_CreateUser_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: &services.FakeUserService{}})
	ctrl := NewUserController(a)

	body := []byte(`{"username":"","password":""}`)
	c, w := newGinContext(http.MethodPost, "/api/users", body)
	ctrl.CreateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid request")
	}
}

func TestUserController_UpdateUser(t *testing.T) {
	userMock := &services.FakeUserService{UpdateUserErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	body, _ := json.Marshal(map[string]interface{}{
		"nickname": "新昵称",
		"password": "",
		"remark":   "新备注",
		"role_ids": []uint{},
	})
	c, w := newGinContextWithParam(http.MethodPut, "/api/users/1", body, "id", "1")
	ctrl.UpdateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}

func TestUserController_DeleteUser(t *testing.T) {
	userMock := &services.FakeUserService{DeleteUserErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	c, w := newGinContextWithParam(http.MethodDelete, "/api/users/1", nil, "id", "1")
	ctrl.DeleteUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}

func TestUserController_ToggleStatus(t *testing.T) {
	userMock := &services.FakeUserService{ToggleStatusErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	c, w := newGinContextWithParam(http.MethodPut, "/api/users/1/status", nil, "id", "1")
	ctrl.ToggleStatus(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}

func TestUserController_ResetPassword(t *testing.T) {
	userMock := &services.FakeUserService{
		ResetPasswordPw:  "newpass123",
		ResetPasswordErr: nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: userMock})
	ctrl := NewUserController(a)

	c, w := newGinContextWithParam(http.MethodPut, "/api/users/1/reset-password", nil, "id", "1")
	ctrl.ResetPassword(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if code, _ := resp["code"].(float64); code != 0 {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
	data, _ := resp["data"].(map[string]interface{})
	if data == nil {
		t.Fatal("expected data")
	}
	if _, ok := data["password"]; !ok {
		t.Error("expected data.password in reset response")
	}
}

func TestUserController_InvalidID(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{UserService: &services.FakeUserService{}})
	ctrl := NewUserController(a)

	c, w := newGinContextWithParam(http.MethodDelete, "/api/users/abc", nil, "id", "abc")
	ctrl.DeleteUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid user id")
	}
}
