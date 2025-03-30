package captcha

type CaptchaConfig struct {
	CaptchaType   string `mapstructure:"captcha_type" json:"captcha_type" yaml:"captcha_type"`
	CaptchaLength int    `mapstructure:"captcha_length" json:"captcha_length" yaml:"captcha_length"`
}
