package errors

import (
	"errors"
	"testing"
)

func TestBizError_Error(t *testing.T) {
	e := &BizError{Code: CodeBadRequest, Msg: "参数错误"}
	if e.Error() != "参数错误" {
		t.Errorf("Error() = %q, want 参数错误", e.Error())
	}
}

func TestNewBizError(t *testing.T) {
	err := NewBizError(CodeUnauthorized, "未认证")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("NewBizError should return *BizError")
	}
	if biz.Code != CodeUnauthorized || biz.Msg != "未认证" {
		t.Errorf("Code=%d Msg=%q, want 401 未认证", biz.Code, biz.Msg)
	}
}

func TestBadRequestErr(t *testing.T) {
	if BadRequestErr(nil) != nil {
		t.Error("BadRequestErr(nil) should return nil")
	}
	err := BadRequestErr(errors.New("invalid json"))
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("BadRequestErr should return *BizError")
	}
	if biz.Code != CodeBadRequest {
		t.Errorf("Code = %d, want %d", biz.Code, CodeBadRequest)
	}
	if biz.Msg != "参数错误: invalid json" {
		t.Errorf("Msg = %q", biz.Msg)
	}
}

func TestBadRequestMsg(t *testing.T) {
	err := BadRequestMsg("缺少必填项")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("BadRequestMsg should return *BizError")
	}
	if biz.Code != CodeBadRequest || biz.Msg != "缺少必填项" {
		t.Errorf("Code=%d Msg=%q", biz.Code, biz.Msg)
	}
}

func TestUnauthorizedMsg(t *testing.T) {
	err := UnauthorizedMsg("Token无效")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("UnauthorizedMsg should return *BizError")
	}
	if biz.Code != CodeUnauthorized || biz.Msg != "Token无效" {
		t.Errorf("Code=%d Msg=%q", biz.Code, biz.Msg)
	}
}

func TestForbiddenMsg(t *testing.T) {
	err := ForbiddenMsg("无权限")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("ForbiddenMsg should return *BizError")
	}
	if biz.Code != CodeForbidden || biz.Msg != "无权限" {
		t.Errorf("Code=%d Msg=%q", biz.Code, biz.Msg)
	}
}

func TestNotFoundMsg(t *testing.T) {
	err := NotFoundMsg("用户不存在")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("NotFoundMsg should return *BizError")
	}
	if biz.Code != CodeNotFound || biz.Msg != "用户不存在" {
		t.Errorf("Code=%d Msg=%q", biz.Code, biz.Msg)
	}
}

func TestInternalErrorMsg(t *testing.T) {
	err := InternalErrorMsg("服务器内部错误")
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatal("InternalErrorMsg should return *BizError")
	}
	if biz.Code != CodeInternalError || biz.Msg != "服务器内部错误" {
		t.Errorf("Code=%d Msg=%q", biz.Code, biz.Msg)
	}
}
