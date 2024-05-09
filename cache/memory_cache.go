package cache

import (
	"time"

	"github.com/loongkirin/gdk/util"
	memCache "github.com/robfig/go-cache"
)

type InMemoryStore struct {
	memCache.Cache
}

func NewInMemoryStore(defaultExpiration time.Duration) *InMemoryStore {
	return &InMemoryStore{*memCache.New(defaultExpiration, time.Minute)}
}

func (ms *InMemoryStore) Get(key string) (string, error) {
	val, found := ms.Cache.Get(key)
	if !found {
		return "", ErrCacheMiss
	}
	v, err := util.Serialize(val)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (ms *InMemoryStore) Set(key string, value interface{}, expires time.Duration) error {
	ms.Cache.Set(key, value, expires)
	return nil
}

func (ms *InMemoryStore) Add(key string, value interface{}, expires time.Duration) error {
	err := ms.Cache.Add(key, value, expires)
	if err == memCache.ErrKeyExists {
		return ErrNotStored
	}
	return err
}

func (ms *InMemoryStore) Replace(key string, value interface{}, expires time.Duration) error {
	if err := ms.Cache.Replace(key, value, expires); err != nil {
		return ErrNotStored
	}
	return nil
}

func (ms *InMemoryStore) Delete(key string) error {
	if found := ms.Cache.Delete(key); !found {
		return ErrCacheMiss
	}
	return nil
}

func (ms *InMemoryStore) Increment(key string, value int64) (int64, error) {
	newValue, err := ms.Cache.Increment(key, uint64(value))
	if err == memCache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return int64(newValue), err
}

func (ms *InMemoryStore) Decrement(key string, value int64) (int64, error) {
	newValue, err := ms.Cache.Decrement(key, uint64(value))
	if err == memCache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return int64(newValue), err
}

func (ms *InMemoryStore) Flush() error {
	ms.Cache.Flush()
	return nil
}
