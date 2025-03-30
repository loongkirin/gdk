package cache

import (
	"errors"
)

var (
	ErrNotStored  = errors.New("cache: item not stored")
	ErrNotSupport = errors.New("cache: not support")
	ErrKeyExists  = errors.New("cache: item already exists")
	ErrCacheMiss  = errors.New("cache: item not found")
)
