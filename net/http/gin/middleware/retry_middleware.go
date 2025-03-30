package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/gin-gonic/gin"
	gdklogger "github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/net/http/response"
)

func Retry(logger gdklogger.Logger, maxRetries uint, retryDelay time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := retry.Do(
			func() error {
				// 执行下一个处理程序
				cCp := c.Copy()
				cCp.Next()

				// 如果没有错误，直接返回
				// if cCp.Writer.Status() == http.StatusOK || cCp.Writer.Status() == http.StatusNotFound {
				// 	return nil
				// }

				// 返回错误
				if cCp.Writer.Status() >= http.StatusInternalServerError {
					return fmt.Errorf("http request failed with status %d", cCp.Writer.Status())
				}
				return nil
			},
			retry.Attempts(maxRetries), // 重试次数
			retry.Delay(retryDelay),    // 重试间隔
			retry.OnRetry(func(n uint, err error) {
				errMsg := fmt.Sprintf("Retry #%d: %s\n", n, err)
				logger.Error("request failed with retry", gdklogger.Fields{
					"error":     err,
					"traceId":   GetTraceID(c),
					"requestId": GetRequestId(c),
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"message":   errMsg,
				})
			}),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, response.NewResponse(response.ERROR, fmt.Sprintf("Request failed after %d retries", maxRetries)))
		}
	}
}
