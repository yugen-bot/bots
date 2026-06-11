package inits

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/api"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// RunHTTP starts the Fiber HTTP server and blocks until ctx is cancelled,
// then shuts down gracefully. Intended to be run in an errgroup goroutine.
func RunHTTP(ctx context.Context, container *di.Container) error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: utils.Logger.Desugar(),
		SkipURIs: []string{
			"/api",
			"/api/monitor",
			"/api/health",
			"/api/metrics",
		},
	}))

	router := app.Group("/api")
	api.AddSharedMiddleware(app)
	api.AddSharedRoutes(app, router, container)

	cfg := container.Get(static.DiConfig).(*config.Config)
	addr := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)

	utils.Logger.Info("Initializing api...")

	errCh := make(chan error, 1)
	go func() { errCh <- app.Listen(addr) }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		utils.Logger.Info("Shutting down HTTP server...")

		return app.ShutdownWithContext(shutdownCtx)
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("api: listen: %w", err)
		}

		return nil
	}
}
