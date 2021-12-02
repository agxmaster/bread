package requestid

import "context"

type contextKey struct{}

var requestidKey = contextKey{}

// ContextWithRequestID returns a new `context.Context` that holds a requestid
func ContextWithRequestID(ctx context.Context, requestid string) context.Context {
	return context.WithValue(ctx, requestidKey, requestid)
}

// FromContext returns the requestid previously associated with `ctx`, or  "" if not found.
func FromContext(ctx context.Context) string {
	val := ctx.Value(requestidKey)
	if rid, ok := val.(string); ok {
		return rid
	}
	return ""
}
