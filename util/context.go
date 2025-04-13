package util

import (
	"context"
	"math/rand"
	"strings"
)

type scopeCtxKey struct{}

// GetCtxWithScope creates a new context with the given scope.
func GetCtxWithScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, scopeCtxKey{}, scope)
}

// GetScopeFromCtx retrieves the scope from the context.
func GetScopeFromCtx(ctx context.Context) (string, bool) {
	val := ctx.Value(scopeCtxKey{})
	scope, ok := val.(string)
	return scope, ok
}

type traceIdCtxKey struct{}

// GetCtxWithTraceId creates a new context with a generated trace ID.
func GetCtxWithTraceId(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceIdCtxKey{}, generateTraceId())
}

// GetTraceIdFromCtx retrieves the trace ID from the context.
func GetTraceIdFromCtx(ctx context.Context) (string, bool) {
	val := ctx.Value(traceIdCtxKey{})
	traceId, ok := val.(string)
	return traceId, ok
}

// generateTraceId generates a random trace ID in the format "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".
func generateTraceId() string {
	sb := strings.Builder{}
	sb.Grow(36)

	for i := 0; i < 36; i++ {
		switch i {
		case 8, 13, 18, 23:
			sb.WriteByte('-')
			continue
		case 14:
			sb.WriteByte('4') // version 4
			continue
		case 19:
			r := rand.Intn(4) + 8 // values 8–b
			sb.WriteByte("89ab"[r-8])
			continue
		default:
			n := rand.Intn(16)
			sb.WriteByte("0123456789abcdef"[n])
		}
	}

	return sb.String()
}
