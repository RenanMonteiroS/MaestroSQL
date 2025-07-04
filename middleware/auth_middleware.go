package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// This middleware checks if exists a session with some 'userEmail' set into the cookie store
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userEmail := session.Get("userEmail")

		if userEmail == nil {
			c.JSON(http.StatusUnauthorized, map[string]any{"msg": "You are not authorized. Try to login"})
			c.Abort()
			return
		}

		c.Next()
	}
}
