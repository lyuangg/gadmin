package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
)

func TestRoleController_GetRoles(t *testing.T) {
	roleMock := &services.FakeRoleService{
		GetRolesList:  []models.Role{},
		GetRolesTotal: 0,
		GetRolesErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: roleMock})
	ctrl := NewRoleController(a)

	c, w := newGinContextGET("/api/roles?page=1&page_size=10")
	ctrl.GetRoles(c)

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

func TestRoleController_CreateRole(t *testing.T) {
	roleMock := &services.FakeRoleService{
		CreateRoleResult: &models.Role{ID: 1, Name: "测试角色", Description: "描述"},
		CreateRoleErr:    nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: roleMock})
	ctrl := NewRoleController(a)

	body, _ := json.Marshal(map[string]string{"name": "测试角色", "description": "描述"})
	c, w := newGinContext(http.MethodPost, "/api/roles", body)
	ctrl.CreateRole(c)

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
	if data["name"] != "测试角色" {
		t.Errorf("create result: %v", data)
	}
}

func TestRoleController_CreateRole_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: &services.FakeRoleService{}})
	ctrl := NewRoleController(a)

	body := []byte(`{"name":""}`)
	c, w := newGinContext(http.MethodPost, "/api/roles", body)
	ctrl.CreateRole(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid request")
	}
}

func TestRoleController_UpdateRole(t *testing.T) {
	roleMock := &services.FakeRoleService{
		UpdateRoleResult: &models.Role{ID: 1, Name: "新角色", Description: "新描述"},
		UpdateRoleErr:    nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: roleMock})
	ctrl := NewRoleController(a)

	body, _ := json.Marshal(map[string]string{"name": "新角色", "description": "新描述"})
	c, w := newGinContextWithParam(http.MethodPut, "/api/roles/1", body, "id", "1")
	ctrl.UpdateRole(c)

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
	if data["name"] != "新角色" {
		t.Errorf("update result: %v", data)
	}
}

func TestRoleController_DeleteRole(t *testing.T) {
	roleMock := &services.FakeRoleService{DeleteRoleErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: roleMock})
	ctrl := NewRoleController(a)

	c, w := newGinContextWithParam(http.MethodDelete, "/api/roles/1", nil, "id", "1")
	ctrl.DeleteRole(c)

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

func TestRoleController_AssignPermissions(t *testing.T) {
	roleMock := &services.FakeRoleService{AssignPermissionsErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: roleMock})
	ctrl := NewRoleController(a)

	body, _ := json.Marshal(map[string]interface{}{"permission_ids": []uint{1}})
	c, w := newGinContextWithParam(http.MethodPut, "/api/roles/1/permissions", body, "id", "1")
	ctrl.AssignPermissions(c)

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

func TestRoleController_AssignPermissions_InvalidRoleID(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{RoleService: &services.FakeRoleService{}})
	ctrl := NewRoleController(a)

	body, _ := json.Marshal(map[string]interface{}{"permission_ids": []uint{}})
	c, w := newGinContextWithParam(http.MethodPut, "/api/roles/abc/permissions", body, "id", "abc")
	ctrl.AssignPermissions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid role id")
	}
}
