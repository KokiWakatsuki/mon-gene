package clients

import "fmt"

// CustomError represents different types of AI client errors
type CustomError struct {
	Type    ErrorType
	Message string
}

type ErrorType int

const (
	ErrorTypeGeneral ErrorType = iota
	ErrorTypeTokenLimit
	ErrorTypeInvalidAPIKey
	ErrorTypeRateLimit
	ErrorTypeModelNotFound
	ErrorTypeQuotaExceeded
)

func (e *CustomError) Error() string {
	return e.Message
}

// IsTokenLimitError checks if the error is related to token limits
func IsTokenLimitError(err error) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == ErrorTypeTokenLimit
	}
	return false
}

// NewTokenLimitError creates a new token limit error
func NewTokenLimitError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeTokenLimit,
		Message: fmt.Sprintf("トークン数が上限を超えています: %s", message),
	}
}

// NewInvalidAPIKeyError creates a new invalid API key error
func NewInvalidAPIKeyError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeInvalidAPIKey,
		Message: fmt.Sprintf("APIキーが無効です: %s", message),
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeRateLimit,
		Message: fmt.Sprintf("レート制限に達しました: %s", message),
	}
}

// NewModelNotFoundError creates a new model not found error
func NewModelNotFoundError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeModelNotFound,
		Message: fmt.Sprintf("指定されたモデルが見つかりません: %s", message),
	}
}

// NewQuotaExceededError creates a new quota exceeded error
func NewQuotaExceededError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeQuotaExceeded,
		Message: fmt.Sprintf("APIクォータが不足しています: %s", message),
	}
}

// NewGeneralError creates a new general error
func NewGeneralError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeGeneral,
		Message: message,
	}
}
