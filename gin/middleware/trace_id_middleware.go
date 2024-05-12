package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

const (
	traceHeaderKey = "x-trace-id"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceHeader := ctx.GetHeader(traceHeaderKey)
		if len(traceHeader) == 0 {
			ctx.Request.Header.Set(traceHeaderKey, util.GenerateId())
		}

		ctx.Next()
	}
}
