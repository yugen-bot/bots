package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"

	"jurien.dev/yugen/hoshi/internal/inits"

	sharedInits "jurien.dev/yugen/shared/inits"
	"jurien.dev/yugen/shared/utils"
)

func main() {
	_ = godotenv.Load() // .env is optional in production environments

	utils.CreateLogger("hoshi")

	defer utils.Shutdown()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	container, err := inits.InitDI()
	if err != nil {
		utils.Logger.Errorf("init DI: %v", err)
		os.Exit(1)
	}
	defer container.DeleteWithSubContainers()

	if err := inits.InitDiscordBot(ctx, &container); err != nil {
		utils.Logger.Errorf("init discord: %v", err)
		os.Exit(1)
	}

	sharedInits.InitCron(&container)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return sharedInits.RunHTTP(gctx, &container)
	})

	utils.Logger.Info("Started hoshi. Stop with CTRL-C...")

	<-ctx.Done()
	utils.Logger.Info("Shutting down...")

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		utils.Logger.Errorf("shutdown: %v", err)
	}

	utils.Logger.Info("Gracefully shut down.")
}
