package utils

import "go.uber.org/zap"

// LogIfErr logs a non-nil error at warn level.
// Use for Discord API send calls where the best we can do is log and continue.
func LogIfErr(log *zap.SugaredLogger, action string, err error) {
	if err != nil {
		log.Warnw("discord action failed", "action", action, "error", err)
	}
}
