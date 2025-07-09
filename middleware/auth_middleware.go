package middleware

import (
	"net/http"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// This middleware checks if exists a session with some 'userEmail' set into the cookie store
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userEmail := session.Get("userEmail")

		if userEmail == nil {
			ctx.JSON(http.StatusUnauthorized, model.APIResponse{Status: "error", Code: http.StatusUnauthorized, Message: "You are not authorized. Try to login", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
