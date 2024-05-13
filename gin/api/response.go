package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

const (
	ERROR        = 500
	UNAUTHORIZED = 401
	BADREQUEST   = 400
	SUCCESS      = 200
)

func NewResponseWithData(code int, msg string, result interface{}) Response {
	return Response{
		code,
		msg,
		result,
	}
}

func NewResponse(code int, msg string) Response {
	return Response{
		code,
		msg,
		map[string]interface{}{},
	}
}

func Result(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, NewResponseWithData(code, msg, data))
}

func Ok(c *gin.Context, msg string, data interface{}) {
	Result(c, SUCCESS, msg, data)
}

func Fail(c *gin.Context, msg string, data interface{}) {
	Result(c, ERROR, msg, data)
}

func Unauthorized(c *gin.Context, msg string, data interface{}) {
	Result(c, UNAUTHORIZED, msg, data)
}

func BadRequest(c *gin.Context, msg string, data interface{}) {
	Result(c, BADREQUEST, msg, data)
}
