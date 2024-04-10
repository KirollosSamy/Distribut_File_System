package utils

import "sync"

type SafeMap[K comparable, V any] struct {
	data map[K]V
	mutx sync.RWMutex
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
    return &SafeMap[K, V]{
        data: make(map[K]V),
    }
}

func (sm* SafeMap[K, V]) Set(key K, value V) {
	sm.mutx.Lock()
	defer sm.mutx.Unlock()
	sm.data[key] = value
}

func (sm* SafeMap[K, V]) Get(key K) V {
	sm.mutx.RLock()
	defer sm.mutx.RUnlock()
	return sm.data[key]
}

// Return a copy of the underlying map
func (sm *SafeMap[K, V]) GetMap() map[K]V {
	sm.mutx.RLock()
	defer sm.mutx.RUnlock()

	copyMap := make(map[K]V)
	for k, v := range sm.data {
		copyMap[k] = v
	}
	return copyMap
}
