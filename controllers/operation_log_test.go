package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/services"
)

func TestOperationLogController_GetOperationLogs(t *testing.T) {
	logMock := &services.FakeOperationLogService{
		GetOperationLogsList:  nil,
		GetOperationLogsTotal: 0,
		GetOperationLogsErr:   nil,
	}
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{OperationLogService: logMock})
	ctrl := NewOperationLogController(a)

	c, w := newGinContextGET("/api/operation-logs?page=1&page_size=10")
	ctrl.GetOperationLogs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.Bytes())
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

func TestOperationLogController_GetOperationLogs_InvalidStartTime(t *testing.T) {
	a := app.NewTestAppWithServiceMocks(&app.ServiceMocks{OperationLogService: &services.FakeOperationLogService{}})
	ctrl := NewOperationLogController(a)

	c, w := newGinContextGET("/api/operation-logs?page=1&page_size=10&start_time=invalid")
	ctrl.GetOperationLogs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if code, _ := resp["code"].(float64); code == 0 {
		t.Error("expected error for invalid start_time")
	}
}
