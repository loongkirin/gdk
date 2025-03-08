package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func RequestIdMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestId := GetRequestId(ctx)
		if len(requestId) == 0 {
			SetRequestId(ctx, util.GenerateId())
		}

		ctx.Next()
	}
}
