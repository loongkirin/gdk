package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/net/http/response"
	"github.com/loongkirin/gdk/oauth"
)

const (
	authorizationHeaderKey  = "x-authorization"
	authorizationTypeBearer = "x-bearer"
	authorizationClaimsKey  = "x-authorization-claims"
)

func OAuth(oauthMaker oauth.OAuthMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if authorizationHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization header invalid"))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization header invalid"))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization type invalid"))
			return
		}

		accessToken := fields[1]
		claims, err := oauthMaker.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, err.Error()))
			return
		}

		c.Set(authorizationClaimsKey, claims)
		c.Next()
	}
}
