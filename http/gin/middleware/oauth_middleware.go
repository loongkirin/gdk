package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/http/response"
	"github.com/loongkirin/gdk/oauth"
)

const (
	authorizationHeaderKey  = "x-authorization"
	authorizationTypeBearer = "x-bearer"
	authorizationClaimsKey  = "x-authorization-claims"
)

func OAuthMiddleware(oauthMaker oauth.OAuthMaker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization header invalid"))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization header invalid"))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, "Authorization type invalid"))
			return
		}

		accessToken := fields[1]
		claims, err := oauthMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewResponse(response.UNAUTHORIZED, err.Error()))
			return
		}

		ctx.Set(authorizationClaimsKey, claims)
		ctx.Next()
	}
}
