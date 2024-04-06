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
	sm.mutx.Lock()
	defer sm.mutx.Unlock()
	return sm.data[key]
}