package apperror

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with an AppError.
func Wrap(err error, code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Common application errors
var (
	ErrDuplicateEmail = &AppError{
		Code:       "DUPLICATE_EMAIL",
		Message:    "An account with this email already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrDuplicateEntry = &AppError{
		Code:       "DUPLICATE_ENTRY",
		Message:    "A record with this value already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "The requested resource was not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Invalid input data",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An unexpected error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrForeignKeyViolation = &AppError{
		Code:       "FOREIGN_KEY_VIOLATION",
		Message:    "Referenced record does not exist",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Postgres error codes
const (
	PgUniqueViolation     = "23505"
	PgForeignKeyViolation = "23503"
	PgNotNullViolation    = "23502"
	PgCheckViolation      = "23514"
)

// MapPostgresError maps a Postgres error to an AppError.
// It checks for specific constraint names to provide more context.
func MapPostgresError(err error, constraintMapping map[string]*AppError) *AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	// Check for specific constraint mappings first
	if constraintMapping != nil {
		if appErr, ok := constraintMapping[pgErr.ConstraintName]; ok {
			return &AppError{
				Code:       appErr.Code,
				Message:    appErr.Message,
				HTTPStatus: appErr.HTTPStatus,
				Err:        err,
			}
		}
	}

	// Fall back to generic error code mapping
	switch pgErr.Code {
	case PgUniqueViolation:
		return Wrap(err, ErrDuplicateEntry.Code, ErrDuplicateEntry.Message, ErrDuplicateEntry.HTTPStatus)
	case PgForeignKeyViolation:
		return Wrap(err, ErrForeignKeyViolation.Code, ErrForeignKeyViolation.Message, ErrForeignKeyViolation.HTTPStatus)
	case PgNotNullViolation:
		return Wrap(err, ErrValidation.Code, "A required field is missing", ErrValidation.HTTPStatus)
	case PgCheckViolation:
		return Wrap(err, ErrValidation.Code, "A field value is out of allowed range", ErrValidation.HTTPStatus)
	default:
		return nil
	}
}
