package utils

import (
	"context"
	"log"
	"os"
	"time"

	zaploki "github.com/paul-milne/zap-loki"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"jurien.dev/yugen/shared/static"
)

const productionEnv = "production"

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
	environment := productionEnv

	if os.Getenv(static.Env) != productionEnv {
		cfg = prettyconsole.NewConfig()
		environment = os.Getenv(static.Env)
	}

	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		log.Panic(err)
	}

	logger.Sugar().Infof("Log level is set to %s", cfg.Level.String())

	if len(os.Getenv(static.EnvLokiHost)) > 0 && environment == productionEnv {
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

	if InitSentry(appName) {
		sentryCore, sentryErr := newSentryCore(zapcore.WarnLevel)
		if sentryErr != nil {
			log.Printf("sentry: failed to create zap core: %v", sentryErr)
		} else {
			logger = logger.WithOptions(zap.WrapCore(func(original zapcore.Core) zapcore.Core {
				return zapcore.NewTee(original, sentryCore)
			}))
		}
	}

	Logger = logger.Sugar()

	return Logger
}

// Shutdown flushes all log sinks (Sentry, zap) and should be deferred in main.
func Shutdown() {
	FlushSentry(2 * time.Second)

	_ = Logger.Sync()
}
