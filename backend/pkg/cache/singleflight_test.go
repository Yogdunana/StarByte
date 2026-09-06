package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlight_Dedup(t *testing.T) {
	f := newFlight()
	var calls int32
	var wg sync.WaitGroup
	const n = 40
	wg.Add(n)
	vals := make([]string, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			vals[i], errs[i] = f.Do("hot", func() (string, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(30 * time.Millisecond)
				return "v", nil
			})
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	for i := 0; i < n; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, "v", vals[i])
	}
}

func TestFlight_ErrorShared(t *testing.T) {
	f := newFlight()
	_, err := f.Do("k", func() (string, error) { return "", assert.AnError })
	assert.Error(t, err)
}
