package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Relation struct {
	mu    sync.Mutex // protects calls
	calls map[string]*call
}

func (r *Relation) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	r.mu.Lock()
	if r.calls == nil {
		r.calls = make(map[string]*call)
	}
	if c, ok := r.calls[key]; ok {
		r.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	r.calls[key] = c
	r.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	r.mu.Lock()
	delete(r.calls, key)
	r.mu.Unlock()

	return c.val, c.err
}
