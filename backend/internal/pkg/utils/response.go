package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// APIResponse is the unified API response envelope.
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	ErrorCode int         `json:"error_code,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
	Cause      error
}

const (
	ErrCodeInvalidRequest           = 400001
	ErrCodeForbidden                = 403001
	ErrCodeNotFound                 = 404001
	ErrCodeInternal                 = 500001
	ErrCodeFileTooLarge             = 400101
	ErrCodeFileExtensionUnsupported = 400102
)

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NewAppError(httpStatus int, code int, message string) error {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message}
}

func WrapAppError(httpStatus int, code int, message string, cause error) error {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message, Cause: cause}
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, APIResponse{Code: httpCode, Message: message})
}

func FailWithCode(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, APIResponse{Code: httpCode, Message: message, ErrorCode: errCode})
}

func FailError(c *gin.Context, fallbackHTTPCode int, err error) {
	if err == nil {
		Fail(c, fallbackHTTPCode, "unknown error")
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		httpCode := appErr.HTTPStatus
		if httpCode <= 0 {
			httpCode = fallbackHTTPCode
		}
		FailWithCode(c, httpCode, appErr.Code, appErr.Message)
		return
	}

	Fail(c, fallbackHTTPCode, err.Error())
}
