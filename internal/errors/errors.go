package errors

import (
	"fmt"
)

// S3Error 结构化的 S3 错误类型
type S3Error struct {
	Operation string
	Bucket    string
	Object    string
	Err       error
}

func (e *S3Error) Error() string {
	if e.Object != "" {
		return fmt.Sprintf("%s 操作失败 (bucket: %s, object: %s): %v", 
			e.Operation, e.Bucket, e.Object, e.Err)
	}
	return fmt.Sprintf("%s 操作失败 (bucket: %s): %v", 
		e.Operation, e.Bucket, e.Err)
}

func (e *S3Error) Unwrap() error {
	return e.Err
}

// NewS3Error 创建新的 S3 错误
func NewS3Error(operation, bucket, object string, err error) *S3Error {
	return &S3Error{
		Operation: operation,
		Bucket:    bucket,
		Object:    object,
		Err:       err,
	}
}

// ConfigError 配置相关错误
type ConfigError struct {
	Field   string
	Value   string
	Reason  string
}

func (e *ConfigError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("配置错误 - %s: '%s' (%s)", e.Field, e.Value, e.Reason)
	}
	return fmt.Sprintf("配置错误 - %s: %s", e.Field, e.Reason)
}

// NewConfigError 创建新的配置错误
func NewConfigError(field, value, reason string) *ConfigError {
	return &ConfigError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Value   string
	Rule    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("验证失败 - %s: '%s' (规则: %s)", e.Field, e.Value, e.Rule)
}

// NewValidationError 创建新的验证错误
func NewValidationError(field, value, rule, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// PathError 路径相关错误
type PathError struct {
	Path      string
	Operation string
	Err       error
}

func (e *PathError) Error() string {
	return fmt.Sprintf("路径错误 - %s 操作失败 (path: %s): %v", 
		e.Operation, e.Path, e.Err)
}

func (e *PathError) Unwrap() error {
	return e.Err
}

// NewPathError 创建新的路径错误
func NewPathError(path, operation string, err error) *PathError {
	return &PathError{
		Path:      path,
		Operation: operation,
		Err:       err,
	}
}
