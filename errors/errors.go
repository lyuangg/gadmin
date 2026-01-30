package errors

type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

func NewBizError(code int, msg string) error {
	return &BizError{Code: code, Msg: msg}
}

func BadRequestErr(err error) error {
	if err == nil {
		return nil
	}
	return NewBizError(CodeBadRequest, "参数错误: "+err.Error())
}

func BadRequestMsg(msg string) error {
	return NewBizError(CodeBadRequest, msg)
}

func UnauthorizedMsg(msg string) error {
	return NewBizError(CodeUnauthorized, msg)
}

// ForbiddenMsg 返回 CodeForbidden 的业务错误
func ForbiddenMsg(msg string) error {
	return NewBizError(CodeForbidden, msg)
}

func NotFoundMsg(msg string) error {
	return NewBizError(CodeNotFound, msg)
}

func InternalErrorMsg(msg string) error {
	return NewBizError(CodeInternalError, msg)
}

// 错误码定义
const (
	CodeSuccess        = 0    // 成功
	CodeBadRequest     = 400  // 请求参数错误
	CodeUnauthorized   = 401  // 未认证
	CodeForbidden      = 403  // 无权限
	CodeNotFound       = 404  // 资源不存在
	CodeInternalError  = 500  // 服务器内部错误
	CodeValidationError = 1001 // 验证错误
	CodeBusinessError  = 1002 // 业务逻辑错误
)
