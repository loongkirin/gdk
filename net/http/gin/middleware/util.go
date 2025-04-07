package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/util"
)

const (
	TraceIdHeaderKey   = "x-trace-id"
	RequestIdHeaderKey = "x-request-id"
)

func GetTraceID(c *gin.Context) string {
	return c.GetHeader(TraceIdHeaderKey)
}

func SetTraceID(c *gin.Context, traceId string) {
	c.Request.Header.Set(TraceIdHeaderKey, traceId)
}

func GetRequestId(c *gin.Context) string {
	return c.GetHeader(RequestIdHeaderKey)
}

func SetRequestId(c *gin.Context, requestId string) {
	c.Request.Header.Set(RequestIdHeaderKey, requestId)
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
