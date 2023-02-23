package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Class struct {
	mu    sync.Mutex // protects calls
	calls map[string]*call
}

func (cls *Class) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	cls.mu.Lock()
	if cls.calls == nil {
		cls.calls = make(map[string]*call)
	}
	if c, ok := cls.calls[key]; ok {
		cls.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	cls.calls[key] = c
	cls.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	cls.mu.Lock()
	delete(cls.calls, key)
	cls.mu.Unlock()

	return c.val, c.err
}
