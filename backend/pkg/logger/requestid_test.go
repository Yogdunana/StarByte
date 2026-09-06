package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestIDFrom_Empty(t *testing.T) {
	assert.Equal(t, "", RequestIDFrom(nil))
	assert.Equal(t, "", RequestIDFrom(context.Background()))
}

func TestWithRequestID_RoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-abc")
	assert.Equal(t, "req-abc", RequestIDFrom(ctx))
	assert.Equal(t, "req-abc", RequestIDFrom(WithRequestID(nil, "req-abc")))
}
