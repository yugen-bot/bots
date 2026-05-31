package utils

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitSentry_NoDSN(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")

	got := InitSentry("test-app", zap.NewNop())
	if got {
		t.Error("InitSentry should return false when SENTRY_DSN is empty")
	}
}

func TestFlushSentry_WhenDisabled(t *testing.T) {
	sentryEnabled = false

	// Should not panic
	FlushSentry(10 * time.Millisecond)
}
