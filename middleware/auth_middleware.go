package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// This middleware checks if exists a session with some 'userEmail' set into the cookie store
func AuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

		userEmail := sess.Get("userEmail")

		if userEmail == nil {
			return ctx.Status(http.StatusUnauthorized).JSON(model.APIResponse{Status: "error", Code: http.StatusUnauthorized, Message: "You are not authorized. Try to login", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

		return ctx.Next()
	}
}

func SessionMiddleware(store *session.Store) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		sess, err := store.Get(ctx)
		if err != nil {
			slog.Error("Error trying to get the session", "Error", err)
			return err
		}
		ctx.Locals("session", sess)
		return ctx.Next()
	}
}
