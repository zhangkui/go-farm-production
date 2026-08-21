package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError carries a stable business error code plus an HTTP status so the
// transport layer can map it to a consistent JSON response without type
// switches. Every Service returns *AppError (wrapped) for business failures.
type AppError struct {
	Code     int    // business error code (see docs error-code ranges)
	HTTP     int    // mapped HTTP status
	Message  string // user-facing message (Chinese)
	Internal error  // wrapped underlying error, if any
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("code=%d msg=%s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("code=%d msg=%s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Internal }

// NewAppError builds an AppError.
func NewAppError(code, httpStatus int, msg string) *AppError {
	return &AppError{Code: code, HTTP: httpStatus, Message: msg}
}

// Wrap wraps an underlying error into an AppError.
func Wrap(code, httpStatus int, msg string, err error) *AppError {
	return &AppError{Code: code, HTTP: httpStatus, Message: msg, Internal: err}
}

// Error code ranges (aligned with design doc §9.4).
const (
	CodeSuccess = 0

	CodeBadRequest     = 40000
	CodeInvalidJSON    = 40001
	CodeValidation     = 40002
	CodeInvalidOrderBy = 40003

	CodeUnauthorized    = 40100
	CodeInvalidCreds    = 40101
	CodeTokenExpired    = 40102
	CodeTokenInvalid    = 40103
	CodeRefreshRevoked  = 40104
	CodeAccountDisabled = 40105

	CodeForbidden  = 40300
	CodePermission = 40301

	CodeNotFound      = 40400
	CodeUserNotFound  = 40401
	CodePlanNotFound  = 40402
	CodeBatchNotFound = 40403

	CodeConflict          = 40900
	CodePlanOverlap       = 40901
	CodeDuplicate         = 40902
	CodeStateTransition   = 40903
	CodeInsufficientStock = 40904
	CodeReturnExceedsUsed = 40905

	CodeBusinessRule   = 42200
	CodeAreaExceeds    = 42201
	CodeWeightMismatch = 42202

	CodeInternal = 50000
	CodeDB       = 50001
	CodeRedis    = 50002
	CodeTx       = 50003
)

// Common sentinel errors.
var (
	ErrNotFound           = NewAppError(CodeNotFound, http.StatusNotFound, "资源不存在")
	ErrPermission         = NewAppError(CodePermission, http.StatusForbidden, "无操作权限")
	ErrUnauthorized       = NewAppError(CodeUnauthorized, http.StatusUnauthorized, "未认证")
	ErrInvalidCredentials = NewAppError(CodeInvalidCreds, http.StatusUnauthorized, "用户名或密码错误")
	ErrAccountDisabled    = NewAppError(CodeAccountDisabled, http.StatusForbidden, "账号已停用")
	ErrTokenInvalid       = NewAppError(CodeTokenInvalid, http.StatusUnauthorized, "令牌无效")
	ErrTokenExpired       = NewAppError(CodeTokenExpired, http.StatusUnauthorized, "令牌已过期")
	ErrRefreshRevoked     = NewAppError(CodeRefreshRevoked, http.StatusUnauthorized, "刷新令牌已撤销")
	ErrValidation         = NewAppError(CodeValidation, http.StatusBadRequest, "请求参数校验失败")
	ErrConflict           = NewAppError(CodeConflict, http.StatusConflict, "资源冲突")
	ErrPlanOverlap        = NewAppError(CodePlanOverlap, http.StatusConflict, "种植计划时间与该地块现有计划重叠")
	ErrDuplicate          = NewAppError(CodeDuplicate, http.StatusConflict, "资源已存在")
	ErrStateTransition    = NewAppError(CodeStateTransition, http.StatusConflict, "状态流转不合法")
	ErrInsufficientStock  = NewAppError(CodeInsufficientStock, http.StatusConflict, "投入品批次剩余数量不足")
	ErrReturnExceedsUsed  = NewAppError(CodeReturnExceedsUsed, http.StatusConflict, "退回数量超过该任务尚未退回的领用数量")
	ErrAreaExceeds        = NewAppError(CodeAreaExceeds, http.StatusUnprocessableEntity, "计划面积超过地块总面积")
	ErrWeightMismatch     = NewAppError(CodeWeightMismatch, http.StatusUnprocessableEntity, "采收明细重量之和不等于总重量")
)

// AsAppError unwraps err to *AppError, returning a generic internal error if
// it is not an AppError.
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return Wrap(CodeInternal, http.StatusInternalServerError, "服务器内部错误", err)
}
