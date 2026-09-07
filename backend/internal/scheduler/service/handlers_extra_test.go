package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler_LookupAndPublic(t *testing.T) {
	RegisterHandler("contract_expiry_test", "合同到期测试", func(_ context.Context, _ string, logf func(string)) error {
		logf("ok")
		return nil
	})
	fn, ok := lookupHandler("contract_expiry_test")
	assert.True(t, ok)
	assert.NotNil(t, fn)
	assert.NoError(t, fn(context.Background(), "", func(string) {}))

	found := false
	for _, h := range publicHandlers() {
		if h.Key == "contract_expiry_test" {
			found = true
			assert.Equal(t, "合同到期测试", h.Description)
		}
	}
	assert.True(t, found)
}
