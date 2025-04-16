package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/loongkirin/gdk/net/http/response"
)

// 自定义验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// 注册自定义验证器
func RegisterCustomValidators(v *validator.Validate) {
	// 手机号验证
	v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) == 11
	})

	// 身份证验证
	v.RegisterValidation("idcard", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) == 18
	})

	// 自定义密码验证（至少包含数字和字母，长度在8-20之间）
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		hasNumber := strings.ContainsAny(value, "0123456789")
		hasLetter := strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		hasSpecial := strings.ContainsAny(value, "!@#$%^&*()_+-=[]{}|;:,.<>?")
		return len(value) >= 8 && len(value) <= 20 && hasNumber && hasLetter && hasSpecial
	})
}

// 注册自定义错误消息
func RegisterCustomMessages() map[string]string {
	return map[string]string{
		"required": "字段是必需的",
		"min":      "值必须大于或等于 %v",
		"max":      "值必须小于或等于 %v",
		"len":      "长度必须等于 %v",
		"email":    "必须是有效的电子邮件地址",
		"mobile":   "必须是有效的手机号",
		"idcard":   "必须是有效的身份证号",
		"password": "密码必须包含数字和字母，长度在8-20之间",
	}
}

// 验证器中间件
func Validator() gin.HandlerFunc {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证器
		RegisterCustomValidators(v)

		// 注册结构体字段名称翻译
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}

	return func(c *gin.Context) {
		c.Next()
	}
}

// 验证请求
func ValidateRequest(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make([]ValidationError, 0)
			messages := RegisterCustomMessages()

			for _, e := range validationErrors {
				field := e.Field()
				tag := e.Tag()
				value := e.Value()

				message := messages[tag]
				if message == "" {
					message = fmt.Sprintf("validate error: %s", tag)
				}

				errors = append(errors, ValidationError{
					Field:   field,
					Tag:     tag,
					Value:   fmt.Sprintf("%v", value),
					Message: message,
				})
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewResponse(response.BADREQUEST, errors[0].Message))
			return err
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, response.NewResponse(response.BADREQUEST, err.Error()))
		return err
	}
	return nil
}
