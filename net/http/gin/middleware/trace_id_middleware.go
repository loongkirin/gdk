package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func TraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := GetOrSetTraceID(c)
		if traceId == "" {
			traceId = util.GenerateId()
			SetTraceID(c, traceId)
		}
		c.Set(TraceIdHeaderKey, traceId)
		c.Writer.Header().Set(TraceIdHeaderKey, traceId)
		c.Next()
	}
}
