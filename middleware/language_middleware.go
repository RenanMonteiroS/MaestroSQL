package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func LanguageMiddleware(bundle *i18n.Bundle) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Gets the Accept Language header
		lang := c.Get("Accept-Language")

		// Starts a new localizer looks up messages in the bundle according to the language preferences in langs
		localizer := i18n.NewLocalizer(bundle, lang)

		// Sets a key-value with the localizer
		c.Locals("localizer", localizer)

		return c.Next()
	}
}
