package processor

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// ValidateError 表示简单的验证错误
type ValidateError struct {
	Message  string      // 错误消息
	Field    string      // 字段名
	Tag      string      // 验证标签
	Value    interface{} // 字段值
	RawError error       // 原始错误
}

// Error 实现error接口
func (e *ValidateError) Error() string {
	return e.Message
}

// Unwrap 返回原始错误
func (e *ValidateError) Unwrap() error {
	return e.RawError
}

// NewValidateError 从validator.FieldError创建ValidateError
func NewValidateError(fieldError validator.FieldError, translator ut.Translator) *ValidateError {
	message := fieldError.Error()
	if translator != nil {
		message = fieldError.Translate(translator)
	}

	return &ValidateError{
		Field:    fieldError.Field(),
		Tag:      fieldError.Tag(),
		Value:    fieldError.Value(),
		Message:  message,
		RawError: fieldError,
	}
}

// 简化的错误类型定义
const (
	// 简化的错误类型导入
	SimpleErrorImport = `
	"github.com/go-playground/validator/v10"
`

	// 简化的错误类型定义字符串（用于生成到validation.go文件中）
	SimpleErrorTypeDefinitionStr = `
// ValidateError 表示简单的验证错误
type ValidateError struct {
	Message  string      // 错误消息
	Field    string      // 字段名
	Tag      string      // 验证标签
	Value    interface{} // 字段值
	RawError error       // 原始错误
}

// Error 实现error接口
func (e *ValidateError) Error() string {
	return e.Message
}

// Unwrap 返回原始错误
func (e *ValidateError) Unwrap() error {
	return e.RawError
}
`
)
