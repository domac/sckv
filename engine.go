package sckv

import (
	"sync"
	"errors"
)

var (
	engineMu sync.RWMutex
	engines  = make(map[string]Engine)
)

type EngineConfig struct {
}

func Register(name string, engine Engine) error {
	engineMu.Lock()
	defer engineMu.Unlock()
	if engine == nil {
		errors.New("engine:register engine is nil")
	}
	if _, dup := engines[name]; dup {
		errors.New("engine: register called twice for cache " + name)
	}
	engines[name] = engine
	return nil
}

//引擎接口
type Engine interface {
	New(cfg *EngineConfig) SCCache
}

func New(engineName string) (SCCache, error) {
	engineMu.RLock()
	engine, ok := engines[engineName]
	engineMu.RUnlock()
	if !ok {
		return nil, errors.New("engine:unknown engine " + engineName + " (forgotten import?)")
	}
	cache := engine.New(&EngineConfig{})
	return cache, nil
}

//缓存接口
type SCCache interface {
	Set(key, val []byte)
	Get(key []byte) []byte
}
