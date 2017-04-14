package ipcrouter

import (
  "sync"
)

type handlers struct {
	Map map[uint32]Handler
	sync.RWMutex
}

func newhandlers() *handlers {
	return &handlers{
		Map: make(map[uint32]Handler),
	}
}

func (t *handlers) get(key uint32) (Handler, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *handlers) set(key uint32, val Handler) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *handlers) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}


