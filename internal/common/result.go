package common

// Result 统一响应结构（对齐 Java Result<T>）。
type Result[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// Ok 成功响应。
func Ok[T any](data T) Result[T] {
	return Result[T]{Code: 0, Message: "success", Data: data}
}

// OkMsg 带消息的成功。
func OkMsg[T any](msg string, data T) Result[T] {
	return Result[T]{Code: 0, Message: msg, Data: data}
}

// Fail 失败响应。
func Fail[T any](code int, msg string) Result[T] {
	return Result[T]{Code: code, Message: msg}
}

// PageResult 分页响应。
type PageResult[T any] struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Rows     []T   `json:"rows"`
}

func Page[T any](total int64, page, pageSize int, rows []T) PageResult[T] {
	return PageResult[T]{Total: total, Page: page, PageSize: pageSize, Rows: rows}
}
