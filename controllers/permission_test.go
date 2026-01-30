package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
)

func TestPermissionController_GetPermissions(t *testing.T) {
	permMock := &services.FakePermissionService{
		GetPermissionsList:  []models.Permission{},
		GetPermissionsTotal: 0,
		GetPermissionsErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: permMock})
	ctrl := NewPermissionController(a)

	c, w := newGinContextGET("/api/permissions?page=1&page_size=10")
	ctrl.GetPermissions(c)

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

func TestPermissionController_CreatePermission(t *testing.T) {
	permMock := &services.FakePermissionService{
		CreatePermissionResult: &models.Permission{ID: 1, Path: "/api/test", Method: "GET", Name: "测试权限", Group: "测试", Description: "描述"},
		CreatePermissionErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: permMock})
	ctrl := NewPermissionController(a)

	body, _ := json.Marshal(map[string]string{
		"path":        "/api/test",
		"method":      "GET",
		"name":        "测试权限",
		"group":       "测试",
		"description": "描述",
	})
	c, w := newGinContext(http.MethodPost, "/api/permissions", body)
	ctrl.CreatePermission(c)

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
	if data["path"] != "/api/test" || data["method"] != "GET" {
		t.Errorf("create result: %v", data)
	}
}

func TestPermissionController_CreatePermission_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: &services.FakePermissionService{}})
	ctrl := NewPermissionController(a)

	body := []byte(`{"path":"","method":""}`)
	c, w := newGinContext(http.MethodPost, "/api/permissions", body)
	ctrl.CreatePermission(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid request")
	}
}

func TestPermissionController_UpdatePermission(t *testing.T) {
	permMock := &services.FakePermissionService{
		UpdatePermissionResult: &models.Permission{ID: 1, Name: "新名", Group: "新组", Description: "新描述"},
		UpdatePermissionErr:    nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: permMock})
	ctrl := NewPermissionController(a)

	body, _ := json.Marshal(map[string]string{"name": "新名", "group": "新组", "description": "新描述"})
	c, w := newGinContextWithParam(http.MethodPut, "/api/permissions/1", body, "id", "1")
	ctrl.UpdatePermission(c)

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
	if data["name"] != "新名" {
		t.Errorf("update result: %v", data)
	}
}

func TestPermissionController_DeletePermission(t *testing.T) {
	permMock := &services.FakePermissionService{DeletePermissionErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: permMock})
	ctrl := NewPermissionController(a)

	c, w := newGinContextWithParam(http.MethodDelete, "/api/permissions/1", nil, "id", "1")
	ctrl.DeletePermission(c)

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

func TestPermissionController_BatchDeletePermissions(t *testing.T) {
	permMock := &services.FakePermissionService{BatchDeletePermissionsErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: permMock})
	ctrl := NewPermissionController(a)

	body, _ := json.Marshal(map[string]interface{}{"ids": []uint{1, 2}})
	c, w := newGinContext(http.MethodPost, "/api/permissions/batch-delete", body)
	ctrl.BatchDeletePermissions(c)

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

func TestPermissionController_BatchDeletePermissions_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{PermissionService: &services.FakePermissionService{}})
	ctrl := NewPermissionController(a)

	body := []byte(`{"ids":[]}`)
	c, w := newGinContext(http.MethodPost, "/api/permissions/batch-delete", body)
	ctrl.BatchDeletePermissions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for empty ids")
	}
}
