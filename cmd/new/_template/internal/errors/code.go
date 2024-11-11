package errors

type ErrorCode int

const (
	// System Error Codes (1000-1499)
	ErrInternal ErrorCode = iota + 1000
	ErrDatabase
)

// Authentication Error Codes (1500-1999)
const (
	ErrUnauthorized ErrorCode = iota + 1500
	ErrForbidden
)

// Validation Error Codes (3000-3999)
const (
	ErrBadRequest ErrorCode = iota + 3000
	ErrInvalidInput
	ErrMissingField
)

// Business Logic Error Codes (4000-4999)
const (
	ErrResourceNotFound ErrorCode = iota + 4000
	ErrDuplicateEntry
	ErrAccountAlreadyExists
)
