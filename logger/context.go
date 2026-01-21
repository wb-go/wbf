package logger

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is a private type used to avoid key collisions in context.WithValue.
type contextKey struct{}

var requestIDKey = contextKey{}

// SetRequestID stores the given request ID in the context.
// This ID is typically used for distributed tracing and log correlation.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present or if the value is not a string.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GenerateRequestID creates a new globally unique identifier (UUID v4) for use as a request ID.
// The generated ID is a standard string representation of a UUID.
func GenerateRequestID() string {
	return uuid.New().String()
}
