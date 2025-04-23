// Package util provides validation utilities and custom validators
// for common validation scenarios.
package util

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// ValidationMessages stores custom validation error messages
var (
	validationMessages sync.Map
	defaultMessageFunc = func(ve validator.FieldError) string {
		return ve.Error()
	}
)

// PasswordValidator validates password complexity requirements
// including numbers, letters, and special characters
var PasswordValidator = NewCustomValidator("password", func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	param := fl.Param()
	matchModels := strings.Split(param, ",")
	for _, matchModel := range matchModels {
		switch matchModel {
		case "number":
			if !strings.ContainsAny(value, "0123456789") {
				return false
			}
		case "letter":
			if !strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				return false
			}
		case "special":
			if !strings.ContainsAny(value, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
				return false
			}
		}
	}
	return true
}, func(ve validator.FieldError) string {
	value := ve.Field()
	param := ve.Param()
	matchModels := strings.Split(param, ",")
	var messages [3]string
	for _, matchModel := range matchModels {
		switch matchModel {
		case "number":
			if !strings.ContainsAny(value, "0123456789") {
				messages[0] = "数字"
			}
		case "letter":
			if !strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				messages[1] = "字母"
			}
		case "special":
			if !strings.ContainsAny(value, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
				messages[2] = "特殊字符"
			}
		}
	}
	var builder strings.Builder
	for _, message := range messages {
		if message != "" {
			if builder.Len() > 0 {
				builder.WriteString(",")
			}
			builder.WriteString(message)
		}
	}
	return fmt.Sprintf("密码必须包含%s", builder.String())
})

// ChineseMobileValidator validates Chinese mobile phone numbers
var ChineseMobileValidator = NewCustomValidator("cn_mobile", func(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return reg.MatchString(mobile)
}, func(ve validator.FieldError) string {
	return "必须是有效的中国大陆手机号"
})

// ChineseIdcardValidator validates Chinese ID card numbers
var ChineseIdcardValidator = NewCustomValidator("cn_idcard", func(fl validator.FieldLevel) bool {
	idcard := fl.Field().String()
	reg := regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[012])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)
	return reg.MatchString(idcard)
}, func(ve validator.FieldError) string {
	return "必须是有效的中国大陆身份证号"
})

// MinLenValidator validates the minimum length of a field
var MinLenValidator = NewCustomValidator("min_len", func(fl validator.FieldLevel) bool {
	param := fl.Param()
	length := len(fl.Field().String())
	minLen, err := strconv.Atoi(param)
	if err != nil {
		return false
	}
	return length >= minLen
}, func(ve validator.FieldError) string {
	return fmt.Sprintf("字段长度必须大于或等于%s", ve.Param())
})

// MaxLenValidator validates the maximum length of a field
var MaxLenValidator = NewCustomValidator("max_len", func(fl validator.FieldLevel) bool {
	param := fl.Param()
	length := len(fl.Field().String())
	maxLen, err := strconv.Atoi(param)
	if err != nil {
		return false
	}
	return length <= maxLen
}, func(ve validator.FieldError) string {
	return fmt.Sprintf("字段长度必须小于或等于%s", ve.Param())
})

func init() {
	defaultMessages := map[string]func(validator.FieldError) string{
		"required": func(ve validator.FieldError) string {
			return "字段是必需的"
		},
		"min": func(ve validator.FieldError) string {
			return fmt.Sprintf("值必须大于或等于%s", ve.Param())
		},
		"max": func(ve validator.FieldError) string {
			return fmt.Sprintf("值必须小于或等于%s", ve.Param())
		},
		"len": func(ve validator.FieldError) string {
			return fmt.Sprintf("字段长度必须等于%s", ve.Param())
		},
		"email": func(ve validator.FieldError) string {
			return "必须是有效的电子邮件地址"
		},
		"mobile": func(ve validator.FieldError) string {
			return "必须是有效的手机号"
		},
		"default": defaultMessageFunc,
	}

	for k, v := range defaultMessages {
		validationMessages.Store(k, v)
	}
}

// CustomValidator represents a custom validator
type CustomValidator struct {
	Tag          string
	ValidateFunc validator.Func
	MessageFunc  func(validator.FieldError) string
}

func NewCustomValidator(tag string, validateFunc validator.Func, messageFunc func(validator.FieldError) string) CustomValidator {
	return CustomValidator{
		Tag:          tag,
		ValidateFunc: validateFunc,
		MessageFunc:  messageFunc,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Param   string `json:"param"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("字段:%s,错误:%s", ve.Field, ve.Message)
}

// ValidationErrors represents a slice of validation errors
type ValidationErrors []ValidationError

func (ves ValidationErrors) Error() string {
	var builder strings.Builder
	for _, ve := range ves {
		builder.WriteString(ve.Error())
		builder.WriteString("\n")
	}
	return builder.String()
}

// RegisterCustomValidators registers custom validators
func RegisterCustomValidators(v *validator.Validate, validators ...CustomValidator) {
	for _, validator := range validators {
		v.RegisterValidation(validator.Tag, validator.ValidateFunc)
		RegisterCustomMessageFunc(validator.Tag, validator.MessageFunc)
	}
}

// RegisterCustomMessageFuncs registers custom validation error message functions
func RegisterCustomMessageFuncs(msg map[string]func(validator.FieldError) string) {
	for k, v := range msg {
		RegisterCustomMessageFunc(k, v)
	}
}

func RegisterCustomMessageFunc(tag string, messageFunc func(validator.FieldError) string) {
	if messageFunc == nil {
		messageFunc = defaultMessageFunc
	}
	validationMessages.Store(tag, messageFunc)
}

// GetCustomMessageFunc gets the validation error message function for a specific tag
func GetCustomMessageFunc(tag string) func(validator.FieldError) string {
	if msg, ok := validationMessages.Load(tag); ok {
		return msg.(func(validator.FieldError) string)
	}
	return defaultMessageFunc
}

// ValidateStruct is a generic function to validate any struct
func ValidateStruct[T any](v *validator.Validate, obj T) error {
	if err := v.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errs := make(ValidationErrors, 0, len(validationErrors))
			for _, e := range validationErrors {
				errs = append(errs, ValidationError{
					Field:   e.Field(),
					Tag:     e.Tag(),
					Value:   fmt.Sprintf("%v", e.Value()),
					Message: GetCustomMessageFunc(e.Tag())(e),
				})
			}
			return errors.New(errs.Error())
		}
		return err
	}
	return nil
}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func NewDefaultValidator() *defaultValidator {
	return &defaultValidator{}
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		if value.Elem().Kind() != reflect.Struct {
			return v.ValidateStruct(value.Elem().Interface())
		}
		return v.validateStruct(obj)
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(ValidationErrors, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err.(ValidationError))
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

// validateStruct receives struct type
func (v *defaultValidator) validateStruct(obj any) error {
	v.lazyinit()
	// return v.validate.Struct(obj)
	return ValidateStruct(v.validate, obj)
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://pkg.go.dev/github.com/go-playground/validator/v10
func (v *defaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
	})
}

func (v *defaultValidator) ValidateStructCtx(ctx context.Context, obj any) error {
	if obj == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return v.ValidateStruct(obj)
	}
}
