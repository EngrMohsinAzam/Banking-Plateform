package domain

import (
	"errors"
	"fmt"
)

// ErrorCode is a stable, machine-readable identifier for domain failures.
// HTTP and gRPC layers map these to status codes without parsing error strings.
type ErrorCode string

const (
	ErrCodeInvalidMoney      ErrorCode = "INVALID_MONEY"
	ErrCodeInvalidIBAN       ErrorCode = "INVALID_IBAN"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeRequestInProgress ErrorCode = "REQUEST_IN_PROGRESS"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeValidation        ErrorCode = "VALIDATION"
)

// DomainError represents an expected business-rule violation.
// Unexpected infrastructure failures should remain plain wrapped errors.
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError constructs a domain error without a wrapped cause.
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// WrapDomainError attaches an underlying cause to a domain error.
func WrapDomainError(code ErrorCode, message string, err error) *DomainError {
	return &DomainError{Code: code, Message: message, Err: err}
}

// AsDomainError reports whether err is (or wraps) a DomainError.
func AsDomainError(err error) (*DomainError, bool) {
	var de *DomainError
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}

// IsDomainCode reports whether err is a DomainError with the given code.
func IsDomainCode(err error, code ErrorCode) bool {
	de, ok := AsDomainError(err)
	return ok && de.Code == code
}
