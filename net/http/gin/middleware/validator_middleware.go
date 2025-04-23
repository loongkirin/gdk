package middleware

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/loongkirin/gdk/util"
)

var (
	validatorOnce sync.Once
)

// Validator middleware initializes the validator
func Validator(validators ...util.CustomValidator) gin.HandlerFunc {
	validatorOnce.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			util.RegisterCustomValidators(v, validators...)
			v.RegisterTagNameFunc(func(fld reflect.StructField) string {
				name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
				if name == "-" {
					return ""
				}
				return name
			})
		}
	})

	return func(c *gin.Context) {
		c.Next()
	}
}

// ValidateRequest validates the request
func ValidateRequest(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errs := make(util.ValidationErrors, 0, len(validationErrors))
			for _, e := range validationErrors {
				message := util.GetCustomMessageFunc(e.Tag())(e)
				errs = append(errs, util.ValidationError{
					Field:   e.Field(),
					Tag:     e.Tag(),
					Param:   e.Param(),
					Value:   fmt.Sprintf("%v", e.Value()),
					Message: message,
				})
			}
			return errors.New(errs.Error())
		}
		return err
	}
	return nil
}
