package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorCode represents different types of errors that can occur in the GraphQL API
type ErrorCode string

const (
	// Authentication errors
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"

	// Validation errors
	ErrorCodeValidation ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidID  ErrorCode = "INVALID_ID"

	// Resource errors
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// System errors
	ErrorCodeInternal ErrorCode = "INTERNAL_ERROR"
)

// GraphQLError represents a custom GraphQL error with additional metadata
type GraphQLError struct {
	Code    ErrorCode
	Message string
	Path    []string
	Fields  map[string]string
}

// Error implements the error interface
func (e *GraphQLError) Error() string {
	return e.Message
}

// NewGraphQLError creates a new GraphQL error
func NewGraphQLError(code ErrorCode, message string) *GraphQLError {
	return &GraphQLError{
		Code:    code,
		Message: message,
		Fields:  make(map[string]string),
	}
}

// WithPath adds a path to the error
func (e *GraphQLError) WithPath(path ...string) *GraphQLError {
	e.Path = path
	return e
}

// WithField adds a field error
func (e *GraphQLError) WithField(field, message string) *GraphQLError {
	e.Fields[field] = message
	return e
}

// ErrorPresenter is a GraphQL error presenter that formats our custom errors
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	if gqlErr, ok := err.(*GraphQLError); ok {
		extensions := map[string]interface{}{
			"code": gqlErr.Code,
		}
		if len(gqlErr.Fields) > 0 {
			extensions["fields"] = gqlErr.Fields
		}

		path := make(ast.Path, len(gqlErr.Path))
		for i, p := range gqlErr.Path {
			path[i] = ast.PathName(p)
		}

		return &gqlerror.Error{
			Message:    gqlErr.Message,
			Path:       path,
			Extensions: extensions,
		}
	}

	// Handle other types of errors
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	return &gqlerror.Error{
		Message: err.Error(),
	}
}

// Helper functions for common error types

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *GraphQLError {
	return NewGraphQLError(ErrorCodeUnauthorized, message)
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *GraphQLError {
	return NewGraphQLError(ErrorCodeForbidden, message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *GraphQLError {
	return NewGraphQLError(ErrorCodeValidation, message)
}

// NewInvalidIDError creates a new invalid ID error
func NewInvalidIDError(id string) *GraphQLError {
	return NewGraphQLError(ErrorCodeInvalidID, fmt.Sprintf("invalid ID: %s", id))
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string, id string) *GraphQLError {
	return NewGraphQLError(ErrorCodeNotFound, fmt.Sprintf("%s not found: %s", strings.Title(resource), id))
}

// NewAlreadyExistsError creates a new already exists error
func NewAlreadyExistsError(resource string, field string, value string) *GraphQLError {
	return NewGraphQLError(ErrorCodeAlreadyExists, fmt.Sprintf("%s with %s '%s' already exists", strings.Title(resource), field, value))
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *GraphQLError {
	return NewGraphQLError(ErrorCodeInternal, message)
}
