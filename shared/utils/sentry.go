package utils

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryzap "github.com/getsentry/sentry-go/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"jurien.dev/yugen/shared/static"
)

var sentryEnabled bool

// InitSentry initialises the Sentry SDK when SENTRY_DSN is set.
// Must be called before CreateLogger so the zapsentry core can be attached.
// Returns true when Sentry is active.
func InitSentry(appName string, logger *zap.Logger) bool {
	dsn := os.Getenv(static.EnvSentryDSN)
	if dsn == "" {
		return false
	}

	environment := os.Getenv(static.EnvSentryEnvironment)
	if environment == "" {
		environment = os.Getenv(static.Env)
	}

	var tracesSampleRate float64

	if raw := os.Getenv(static.EnvSentryTracesSampleRate); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			tracesSampleRate = v
		}
	}

	debug := os.Getenv(static.EnvSentryDebug) == "true"

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		TracesSampleRate: tracesSampleRate,
		Debug:            debug,
		ServerName:       appName,
		AttachStacktrace: true,
		EnableLogs:       true,
	})
	if err != nil {
		logger.Sugar().Errorw("sentry: init failed: %w", err)
		return false
	}

	sentryEnabled = true

	logger.Sugar().
		Infow("Sentry initialised", "env", environment, "app", appName)

	return true
}

// FlushSentry drains the Sentry event queue. Safe to call when Sentry is disabled.
func FlushSentry(timeout time.Duration) {
	if sentryEnabled {
		sentry.Flush(timeout)
	}
}

// newSentryCore returns a zapcore.Core that forwards Warn-and-above to Sentry.
func newSentryCore() *sentryzap.SentryCore {
	ctx := context.Background()

	return sentryzap.NewSentryCore(ctx, sentryzap.Option{
		Level: []zapcore.Level{
			zapcore.InfoLevel,
			zapcore.WarnLevel,
			zapcore.ErrorLevel,
		},
		AddCaller: true,
	})
}
