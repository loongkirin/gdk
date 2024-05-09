package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

type Captcha struct {
	store base64Captcha.Store
}

func NewCaptcha(store base64Captcha.Store) *Captcha {
	return &Captcha{
		store: store,
	}
}

func (c *Captcha) GenerateStringCaptcha() (id, b64s, answer string, err error) {
	driver := base64Captcha.NewDriverString(46, 140, 2, 2, 4, "234567890abcdefghjkmnpqrstuvwxyz", &color.RGBA{240, 240, 246, 246}, nil, []string{"wqy-microhei.ttc"})
	cap := base64Captcha.NewCaptcha(driver, c.store)
	return cap.Generate()
}

func (c *Captcha) GenerateDigitCaptcha() (id, b64s, answer string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	cap := base64Captcha.NewCaptcha(driver, c.store)
	return cap.Generate()
}

func (c *Captcha) Verify(id, code string, clear bool) bool {
	return c.store.Verify(id, code, clear)
}
