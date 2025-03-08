package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceId := GetOrSetTraceID(ctx)
		if len(traceId) == 0 {
			SetTraceID(ctx, util.GenerateId())
		}

		ctx.Next()
	}
}
