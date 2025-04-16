package captcha

import (
	"errors"

	"github.com/mojocn/base64Captcha"
)

var (
	ErrCaptchaExpired  = errors.New("captcha was expired")
	ErrCaptchaRequired = errors.New("captcha is required")
	ErrCaptchaInvalid  = errors.New("captcha is invalid")
)

func VerifyCaptcha(store base64Captcha.Store, id, answer string, clear bool) (bool, error) {
	if store == nil {
		return false, errors.New("parameter error")
	}
	if id == "" {
		return false, ErrCaptchaRequired
	}

	captcha := store.Get(id, clear)
	if captcha == "" {
		return false, ErrCaptchaExpired
	}

	return captcha == answer, nil
}
