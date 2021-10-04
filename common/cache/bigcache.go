package cache

import (
	"github.com/allegro/bigcache/v3"
	"github.com/xiusin/pine"
	"github.com/xiusin/pine/cache"
)

type PineBigCache struct {
	*bigcache.BigCache
	cfg bigcache.Config
}

func New(cfg bigcache.Config) *PineBigCache {
	bigCache, err := bigcache.NewBigCache(cfg)
	if err != nil {
		panic(err)
	}
	pineBigCache := &PineBigCache{cfg: cfg, BigCache: bigCache}
	pine.RegisterOnInterrupt(func() {
		pine.Logger().Debug("Close bigcache handle")
		pineBigCache.Close()
	})
	return pineBigCache
}

func (r *PineBigCache) Get(key string) ([]byte, error) {
	return r.BigCache.Get(key)
}

func (r *PineBigCache) GetWithUnmarshal(key string, receiver interface{}) error {
	byts, err := r.Get(key)
	if err == nil {
		err = cache.UnMarshal(byts, receiver)
	} else if err == bigcache.ErrEntryNotFound {
		err = cache.ErrKeyNotFound
	}
	return err
}

func (r *PineBigCache) Set(key string, val []byte, ttl ...int) error {
	return r.BigCache.Set(key, val)
}

func (r *PineBigCache) SetWithMarshal(key string, structData interface{}, ttl ...int) error {
	data, err := cache.Marshal(structData)
	if err != nil && err != bigcache.ErrEntryNotFound {
		return err
	}
	return r.Set(key, data, ttl...)
}

func (r *PineBigCache) Delete(key string) error {
	return r.BigCache.Delete(key)
}

func (r *PineBigCache) Remember(key string, receiver interface{}, call func() (interface{}, error), ttl ...int) error {
	var err error
	if err = r.GetWithUnmarshal(key, receiver); err == nil || err == bigcache.ErrEntryNotFound {
		if receiver, err = call(); err == nil {
			err = r.SetWithMarshal(key, receiver, ttl...)
		}
	}
	return err
}

func (r *PineBigCache) Exists(key string) bool {
	byts, _ := r.BigCache.Get(key)
	return byts != nil
}
