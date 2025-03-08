package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

const (
	traceHeaderKey     = "x-trace-id"
	requestIdHeaderKey = "x-request-id"
)

func GetTraceID(c *gin.Context) string {
	traceId := c.GetHeader(traceHeaderKey)
	return traceId
}

func SetTraceID(c *gin.Context, traceId string) {
	c.Header(traceHeaderKey, traceId)
}

func GetRequestId(c *gin.Context) string {
	requestId := c.GetHeader(requestIdHeaderKey)
	return requestId
}

func SetRequestId(c *gin.Context, requestId string) {
	c.Header(requestIdHeaderKey, requestId)
}

func GetOrSetTraceID(c *gin.Context) string {
	traceId := GetTraceID(c)
	if traceId == "" {
		traceId = util.GenerateId()
		SetTraceID(c, traceId)
	}
	return traceId
}

func GetOrSetRequestId(c *gin.Context) string {
	requestId := GetRequestId(c)
	if requestId == "" {
		requestId = util.GenerateId()
		SetRequestId(c, requestId)
	}
	return requestId
}
