package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func TraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := GetOrSetTraceID(c)
		if len(traceId) == 0 {
			SetTraceID(c, util.GenerateId())
		}

		c.Next()
	}
}
