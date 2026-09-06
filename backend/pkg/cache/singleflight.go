package cache

import "sync"

type sfCall struct {
	wg  sync.WaitGroup
	val string
	err error
}

// flight deduplicates concurrent loads for the same key (cache stampede / 击穿).
type flight struct {
	mu sync.Mutex
	m  map[string]*sfCall
}

func newFlight() *flight {
	return &flight{m: map[string]*sfCall{}}
}

func (f *flight) Do(key string, fn func() (string, error)) (string, error) {
	f.mu.Lock()
	if c, ok := f.m[key]; ok {
		f.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &sfCall{}
	c.wg.Add(1)
	f.m[key] = c
	f.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	f.mu.Lock()
	delete(f.m, key)
	f.mu.Unlock()
	return c.val, c.err
}
