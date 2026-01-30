package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/models"
	"github.com/lyuangg/gadmin/services"
)

func TestDictionaryController_GetTypes(t *testing.T) {
	dictMock := &services.FakeDictionaryService{
		GetTypesList:  []models.DictType{},
		GetTypesTotal: 0,
		GetTypesErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{DictionaryService: dictMock})
	ctrl := NewDictionaryController(a)

	c, w := newGinContextGET("/api/dict/types?page=1&page_size=10")
	ctrl.GetTypes(c)

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
	if _, ok := data["data"]; !ok {
		t.Error("expected data.data (list)")
	}
	if _, ok := data["pagination"]; !ok {
		t.Error("expected data.pagination")
	}
}

func TestDictionaryController_CreateType(t *testing.T) {
	dictMock := &services.FakeDictionaryService{
		CreateTypeResult: &models.DictType{ID: 1, Code: "test_type", Name: "测试类型", Remark: "备注"},
		CreateTypeErr:    nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{DictionaryService: dictMock})
	ctrl := NewDictionaryController(a)

	body, _ := json.Marshal(map[string]string{
		"code":   "test_type",
		"name":   "测试类型",
		"remark": "备注",
	})
	c, w := newGinContext(http.MethodPost, "/api/dict/types", body)
	ctrl.CreateType(c)

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
	if data["code"] != "test_type" || data["name"] != "测试类型" {
		t.Errorf("create result: %v", data)
	}
}

func TestDictionaryController_CreateType_BadRequest(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{DictionaryService: &services.FakeDictionaryService{}})
	ctrl := NewDictionaryController(a)

	body := []byte(`{"code":"","name":""}`) // 空 code/name
	c, w := newGinContext(http.MethodPost, "/api/dict/types", body)
	ctrl.CreateType(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid request")
	}
}

func TestDictionaryController_DeleteType(t *testing.T) {
	dictMock := &services.FakeDictionaryService{DeleteTypeErr: nil}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{DictionaryService: dictMock})
	ctrl := NewDictionaryController(a)

	c, w := newGinContextWithParam(http.MethodDelete, "/api/dict/types/1", nil, "id", "1")
	ctrl.DeleteType(c)

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
