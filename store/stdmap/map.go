package stdmap

import (
	"github.com/domac/sckv"
)

type start struct{}

func (s *start) New(config *sckv.EngineConfig) (sckv.SCCache) {
	return &mapCache{
		mc: make(map[string]*bucket),
	}
}

func init() {
	sckv.Register("MapCache", &start{})
}

//继承SCCache接口
type mapCache struct {
	// map cache
	mc map[string]*bucket
}

type bucket struct {
	//Expire time
	expire int64

	//value
	val []byte
}

func (m *mapCache) Set(key, val []byte) {
	m.mc[string(key)] = &bucket{expire: 0, val: val}
}

func (m *mapCache) Get(key []byte) []byte {
	return []byte(m.mc[string(key)].val)
}

func (m *mapCache) del(key []byte) {
	delete(m.mc, string(key))
}
