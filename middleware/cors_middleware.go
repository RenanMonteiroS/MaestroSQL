package middleware

import (
	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     config.AppCORSAllowOrigins,
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type, Authorization, Accept-Language, X-Csrf-Token",
		AllowCredentials: true,
	})
}
