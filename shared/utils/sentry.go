package utils

import (
	"log"
	"os"
	"strconv"
	"time"

	zapsentry "github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
	"jurien.dev/yugen/shared/static"
)

var sentryEnabled bool

// InitSentry initialises the Sentry SDK when SENTRY_DSN is set.
// Must be called before CreateLogger so the zapsentry core can be attached.
// Returns true when Sentry is active.
func InitSentry(appName string) bool {
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
		log.Printf("sentry: init failed: %v", err)
		return false
	}

	sentryEnabled = true

	log.Printf("sentry: initialised (env=%s, app=%s)", environment, appName)

	return true
}

// FlushSentry drains the Sentry event queue. Safe to call when Sentry is disabled.
func FlushSentry(timeout time.Duration) {
	if sentryEnabled {
		sentry.Flush(timeout)
	}
}

// newSentryCore returns a zapcore.Core that forwards Warn-and-above to Sentry.
func newSentryCore(minLevel zapcore.Level) (zapcore.Core, error) {
	cfg := zapsentry.Configuration{
		Level:             minLevel,
		EnableBreadcrumbs: true,
		BreadcrumbLevel:   zapcore.InfoLevel,
		Tags:              map[string]string{},
	}

	return zapsentry.NewCore(
		cfg,
		zapsentry.NewSentryClientFromClient(sentry.CurrentHub().Client()),
	)
}
