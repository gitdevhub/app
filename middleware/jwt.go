package middleware

import (
	"app/service/jwt"
	"app/service/logger"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
	"strings"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtAuthentication := jwt.InitJWTAuthentication()

		accessDetails, err := jwtAuthentication.ExtractAccessTokenFromRequest(c.Request)
		if err != nil {
			logger.Error("You need to be authorized to access this route")
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You need to be authorized to access this route"})
			return
		}

		userId, err := jwt.GetRedisJWTAuth(accessDetails)
		if err != nil {
			logger.Error("You need to be authorized to access this route 2")
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You need to be authorized to access this route 2"})
			return
		}

		c.Set("userID", userId)
		c.Next()
	}
}
