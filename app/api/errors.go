package api

import (
	"fmt"
	"net/http"

	json "github.com/json-iterator/go"
)

var (
	// ErrInternal is an internal server error.
	ErrInternal = NewGenericError(http.StatusInternalServerError, "Internal server error.")
)

// Error defines API error.
type Error struct {
	Code int
	Data map[string]interface{}
}

// MarshalJSON returns JSON encoding of Error.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Data)
}

func (e *Error) Error() string {
	if detail, ok := e.Data["detail"]; ok && len(e.Data) == 1 {
		return detail.(string)
	}

	d, err := json.Marshal(e.Data)
	if err == nil {
		return string(d)
	}

	return fmt.Sprintf("Error %d", e.Code)
}

// NewError creates new error.
func NewError(code int, m map[string]interface{}) *Error {
	return &Error{
		Code: code,
		Data: m,
	}
}

// NewGenericError creates new generic error.
func NewGenericError(code int, detail string) *Error {
	return &Error{
		Code: code,
		Data: map[string]interface{}{"detail": detail},
	}
}

// NewNotFoundError creates new not found error.
func NewNotFoundError(verboser Verboser) *Error {
	return NewGenericError(http.StatusNotFound, fmt.Sprintf("%s was not found.", verboser.VerboseName()))
}

// NewBadRequestError creates new bad request error.
func NewBadRequestError(detail string) *Error {
	return NewGenericError(http.StatusBadRequest, detail)
}

// NewPermissionDeniedError creates new permission denied error.
func NewPermissionDeniedError() *Error {
	return NewGenericError(http.StatusForbidden, "You do not have permission to perform this action.")
}

// NewRevisionMismatchError creates new revision mismatch error.
func NewRevisionMismatchError(expected, current int) *Error {
	return &Error{
		Code: http.StatusBadRequest,
		Data: map[string]interface{}{"expected_revision": fmt.Sprintf("Revision mismatch. Expected %d, got %d.", expected, current)},
	}
}

// NewCountExceededError creates new count exceeded error.
func NewCountExceededError(name string, limit int) *Error {
	return NewGenericError(http.StatusBadRequest, fmt.Sprintf("%s count exceeded (%d).", name, limit))
}
