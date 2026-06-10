package utils

import (
	"errors"
	"net/http"

	"github.com/disgoorg/disgo/rest"
	"go.uber.org/zap"
)

// LogIfErr logs a non-nil error at warn level.
// Use for Discord API send calls where the best we can do is log and continue.
func LogIfErr(log *zap.SugaredLogger, action string, err error) {
	if err != nil {
		log.Warnw("discord action failed", "action", action, "error", err)
	}
}

// LogIfErrNoRateLimit is like LogIfErr but silently drops HTTP 429 rate-limit
// responses. Use when firing multiple reactions on a message where hitting the
// reaction rate limit is expected and not actionable.
func LogIfErrNoRateLimit(log *zap.SugaredLogger, action string, err error) {
	if err == nil {
		return
	}
	var restErr *rest.Error
	if errors.As(err, &restErr) && restErr.Response != nil &&
		restErr.Response.StatusCode == http.StatusTooManyRequests {
		return
	}
	log.Warnw("discord action failed", "action", action, "error", err)
}
