package captcha

import (
	"time"

	"github.com/loongkirin/gdk/cache"
	"github.com/mojocn/base64Captcha"
)

type captchaStore struct {
	cache      cache.CacheStore
	expiration time.Duration
}

func NewCaptchaStore(cache cache.CacheStore, expiration time.Duration) base64Captcha.Store {
	s := new(captchaStore)
	s.cache = cache
	s.expiration = expiration
	return s
}

func (s *captchaStore) Set(id, value string) error {
	return s.cache.Set(id, value, s.expiration)
}

func (s *captchaStore) Get(id string, clear bool) string {
	v, err := s.cache.Get(id)
	if err == nil {
		if clear {
			_ = s.cache.Delete(id)
		}
		return v
	}
	return ""
}

func (s *captchaStore) Verify(id, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}
