package logger

import "context"

type ctxKey string

const requestIDKey ctxKey = "request_id"

// WithRequestID stores the request id on ctx so GORM and other libraries
// that only see context.Context (not gin.Context) can still recover it.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom returns the request id stored by WithRequestID, or "".
func RequestIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}
