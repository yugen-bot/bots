package utils

import (
	"context"
	"log"
	"os"
	"time"

	zaploki "github.com/paul-milne/zap-loki"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"jurien.dev/yugen/shared/static"
)

var Logger *zap.SugaredLogger

func CreateLogger(appName string) *zap.SugaredLogger {
	lvl := zap.NewAtomicLevel()

	switch os.Getenv(static.EnvLogLevel) {
	case "DEBUG":
		lvl.SetLevel(zap.DebugLevel)
	case "WARN":
		lvl.SetLevel(zap.WarnLevel)
	case "ERROR":
		lvl.SetLevel(zap.ErrorLevel)
	}

	cfg := zap.NewProductionConfig()
	environment := "production"

	if os.Getenv(static.Env) != "production" {
		// cfg = zap.NewDevelopmentConfig()
		cfg = prettyconsole.NewConfig()
		environment = os.Getenv(static.Env)
	}

	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		log.Panic(err)
	}

	logger.Sugar().Infof("Log level is set to %s", cfg.Level.String())

	if err != nil {
		log.Panic(err)
	}

	if len(os.Getenv(static.EnvLokiHost)) > 0 && environment == "production" {
		logger.Info("Received loki endpoint, setting up hook...")
		logger.Debug(os.Getenv(static.EnvLokiHost))

		loki := zaploki.New(context.Background(), zaploki.Config{
			Url:          os.Getenv(static.EnvLokiHost),
			BatchMaxSize: 1000,
			BatchMaxWait: 10 * time.Second,
			Labels: map[string]string{
				"app":         appName,
				"environment": environment,
			},
			Username: os.Getenv(static.EnvLokiUsername),
			Password: os.Getenv(static.EnvLokiPassword),
		})

		logger, err = loki.WithCreateLogger(cfg)
	}

	if err != nil {
		log.Panic(err)
	}

	Logger = logger.Sugar()

	return Logger
}
