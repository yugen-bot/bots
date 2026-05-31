package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/api/handlers"
)

func AddSharedRoutes(
	app *fiber.App,
	router fiber.Router,
	container *di.Container,
) {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	router.Get("/monitor", monitor.New())
	router.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Custom shared handlers
	vote := handlers.GetVoteHandler(container)
	vote.AddRoutes(app, router)
}
