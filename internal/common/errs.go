package common

// 统一错误码（与 ykt-aisaas 保持一致范围）。
const (
	CodeSuccess       = 0
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeTokenInvalid  = 40101
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeConflict      = 40900
	CodeInternal      = 50000
	CodeUpstreamError = 50002
)

// Err 业务错误。
type Err struct {
	Code    int
	Message string
}

func (e *Err) Error() string { return e.Message }

func NewErr(code int, msg string) *Err { return &Err{Code: code, Message: msg} }

// 预定义。
var (
	ErrUnauthorized = NewErr(CodeUnauthorized, "未登录或登录已过期")
	ErrTokenInvalid = NewErr(CodeTokenInvalid, "token 无效")
	ErrForbidden    = NewErr(CodeForbidden, "无权访问")
	ErrNotFound     = NewErr(CodeNotFound, "资源不存在")
	ErrInternal     = NewErr(CodeInternal, "系统内部错误")
)
