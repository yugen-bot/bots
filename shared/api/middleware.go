package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func AddSharedMiddleware(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,HEAD,OPTIONS",
	}))
	app.Use(cache.New())
	app.Use(compress.New())
	app.Use(helmet.New())
	app.Use(idempotency.New())
	app.Use(logger.New())
	app.Use(healthcheck.New())
}
