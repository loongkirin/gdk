package memeorycache

import (
	"time"

	"github.com/loongkirin/gdk/cache"
	"github.com/loongkirin/gdk/util"
)

type InMemeoryStore struct {
	cache *MemeoryCache
}

func NewInMemoryStore(defaultExpiration time.Duration) *InMemeoryStore {
	return &InMemeoryStore{NewMemeoryCache(defaultExpiration, time.Minute)}
}

func (ms *InMemeoryStore) Get(key string) (string, error) {
	val, found := ms.cache.Get(key)
	if !found {
		return "", ErrCacheMiss
	}
	v, err := util.Serialize(val)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (ms *InMemeoryStore) Set(key string, value interface{}, expires time.Duration) error {
	ms.cache.Set(key, value, expires)
	return nil
}

func (ms *InMemeoryStore) Add(key string, value interface{}, expires time.Duration) error {
	err := ms.cache.Add(key, value, expires)
	if err == ErrKeyExists {
		return cache.ErrKeyExists
	}
	return err
}

func (ms *InMemeoryStore) Replace(key string, value interface{}, expires time.Duration) error {
	if err := ms.cache.Replace(key, value, expires); err != nil {
		return cache.ErrNotStored
	}
	return nil
}

func (ms *InMemeoryStore) Delete(key string) error {
	if found := ms.cache.Delete(key); !found {
		return ErrCacheMiss
	}
	return nil
}

func (ms *InMemeoryStore) Increment(key string, value int64) (int64, error) {
	newValue, err := ms.cache.Increment(key, uint64(value))
	if err == ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return int64(newValue), err
}

func (ms *InMemeoryStore) Decrement(key string, value int64) (int64, error) {
	newValue, err := ms.cache.Decrement(key, uint64(value))
	if err == ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return int64(newValue), err
}

func (ms *InMemeoryStore) Flush() error {
	ms.cache.Flush()
	return nil
}
