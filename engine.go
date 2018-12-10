package sckv

type StoreConfig struct {
}

//缓存存储接口
type StoreKV interface {
	New(cfg StoreConfig)
	Set(key, val []byte)
	Get(key []byte) []byte
}
