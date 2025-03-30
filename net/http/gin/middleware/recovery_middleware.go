package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gdklogger "github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/net/http/response"
)

func Recovery(logger gdklogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", gdklogger.Fields{
					"error":     err,
					"traceId":   GetTraceID(c),
					"requestId": GetRequestId(c),
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"message":   err,
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewResponse(response.ERROR, "Internal server error"))
			}
		}()
		c.Next()
	}
}
