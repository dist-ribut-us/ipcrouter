package ipcrouter

import (
  "sync"
)

type callbacks struct {
	Map map[uint32]ResponseCallback
	sync.RWMutex
}

func newcallbacks() *callbacks {
	return &callbacks{
		Map: make(map[uint32]ResponseCallback),
	}
}

func (t *callbacks) get(key uint32) (ResponseCallback, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *callbacks) set(key uint32, val ResponseCallback) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *callbacks) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}

type netcallbacks struct {
	Map map[uint32]NetResponseCallback
	sync.RWMutex
}

func newnetcallbacks() *netcallbacks {
	return &netcallbacks{
		Map: make(map[uint32]NetResponseCallback),
	}
}

func (t *netcallbacks) get(key uint32) (NetResponseCallback, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *netcallbacks) set(key uint32, val NetResponseCallback) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *netcallbacks) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}

type queryServices struct {
	Map map[uint32]QueryService
	sync.RWMutex
}

func newqueryServices() *queryServices {
	return &queryServices{
		Map: make(map[uint32]QueryService),
	}
}

func (t *queryServices) get(key uint32) (QueryService, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *queryServices) set(key uint32, val QueryService) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *queryServices) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}

type commandServices struct {
	Map map[uint32]CommandService
	sync.RWMutex
}

func newcommandServices() *commandServices {
	return &commandServices{
		Map: make(map[uint32]CommandService),
	}
}

func (t *commandServices) get(key uint32) (CommandService, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *commandServices) set(key uint32, val CommandService) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *commandServices) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}

type netQueryServices struct {
	Map map[uint32]NetQueryService
	sync.RWMutex
}

func newnetQueryServices() *netQueryServices {
	return &netQueryServices{
		Map: make(map[uint32]NetQueryService),
	}
}

func (t *netQueryServices) get(key uint32) (NetQueryService, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *netQueryServices) set(key uint32, val NetQueryService) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *netQueryServices) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}

type netCommandServices struct {
	Map map[uint32]NetCommandService
	sync.RWMutex
}

func newnetCommandServices() *netCommandServices {
	return &netCommandServices{
		Map: make(map[uint32]NetCommandService),
	}
}

func (t *netCommandServices) get(key uint32) (NetCommandService, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *netCommandServices) set(key uint32, val NetCommandService) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *netCommandServices) delete(keys ...uint32) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}


