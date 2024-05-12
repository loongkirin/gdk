package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

const (
	requestIdHeaderKey = "x-request-id"
)

func RequestIdMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestIdHeader := ctx.GetHeader(requestIdHeaderKey)
		if len(requestIdHeader) == 0 {
			ctx.Request.Header.Set(requestIdHeaderKey, util.GenerateId())
		}

		ctx.Next()
	}
}
