package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := GetRequestId(c)
		if len(requestId) == 0 {
			SetRequestId(c, util.GenerateId())
		}

		c.Next()
	}
}
