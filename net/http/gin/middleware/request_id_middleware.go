package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := GetOrSetRequestId(c)
		if requestId == "" {
			requestId = util.GenerateId()
			SetRequestId(c, requestId)
		}
		c.Set(RequestIdHeaderKey, requestId)
		c.Writer.Header().Set(RequestIdHeaderKey, requestId)
		c.Next()
	}
}
