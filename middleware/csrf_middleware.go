package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/utils"
)

func CsrfMiddleware() fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-Csrf-Token",
		CookieName:     "csrf_",
		CookieSameSite: "Strict",
		Expiration:     30 * time.Minute,
		KeyGenerator:   utils.UUID,
		ContextKey:     "csrf",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			slog.Error("CSRF Token mismatch.", "Path", ctx.Path(), "Origin", ctx.IP(), "Error", err)
			return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Security validation failure (CSRF). Suspicious request - please check that you are accessing this site correctly.", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		},
	})
}
