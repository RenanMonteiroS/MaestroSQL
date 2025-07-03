package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userEmail := session.Get("userEmail")

		if userEmail == nil {
			c.JSON(http.StatusUnauthorized, map[string]string{"msg": "You are not authorized. Try to login"})
		}

		c.Next()
	}
}
