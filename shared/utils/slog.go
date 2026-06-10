package utils

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewSlogFromZap wraps Yugen's zap logger as an *slog.Logger for use with disgo.
// The returned logger inherits the full zap pipeline (Loki, Sentry, level filter).
func NewSlogFromZap(z *zap.SugaredLogger) *slog.Logger {
	return slog.New(&zapSlogHandler{core: z.Desugar().Core()})
}

type zapSlogHandler struct {
	core   zapcore.Core
	fields []zap.Field
}

func (h *zapSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.core.Enabled(slogToZapLevel(level))
}

func (h *zapSlogHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.core.Enabled(slogToZapLevel(r.Level)) {
		return nil
	}

	fields := make([]zap.Field, 0, len(h.fields)+r.NumAttrs())
	fields = append(fields, h.fields...)

	r.Attrs(func(a slog.Attr) bool {
		fields = append(fields, attrToField(a))
		return true
	})

	entry := zapcore.Entry{
		Level:   slogToZapLevel(r.Level),
		Time:    r.Time,
		Message: r.Message,
	}

	return h.core.Write(entry, fields)
}

func (h *zapSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zap.Field, 0, len(h.fields)+len(attrs))

	fields = append(fields, h.fields...)
	for _, a := range attrs {
		fields = append(fields, attrToField(a))
	}

	return &zapSlogHandler{core: h.core, fields: fields}
}

func (h *zapSlogHandler) WithGroup(name string) slog.Handler {
	// Groups are uncommon in disgo; flatten by prefixing the group name.
	// This keeps the implementation simple without a namespace stack.
	return &zapSlogHandler{
		core: h.core,
		fields: append(
			append([]zap.Field{}, h.fields...),
			zap.String("group", name),
		),
	}
}

func slogToZapLevel(l slog.Level) zapcore.Level {
	switch {
	case l >= slog.LevelError:
		return zapcore.ErrorLevel
	case l >= slog.LevelWarn:
		return zapcore.WarnLevel
	case l >= slog.LevelInfo:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

func attrToField(a slog.Attr) zap.Field {
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindBool:
		return zap.Bool(a.Key, v.Bool())
	case slog.KindInt64:
		return zap.Int64(a.Key, v.Int64())
	case slog.KindUint64:
		return zap.Uint64(a.Key, v.Uint64())
	case slog.KindFloat64:
		return zap.Float64(a.Key, v.Float64())
	case slog.KindString:
		return zap.String(a.Key, v.String())
	case slog.KindTime:
		return zap.Time(a.Key, v.Time())
	case slog.KindDuration:
		return zap.Duration(a.Key, v.Duration())
	case slog.KindGroup:
		m := make(map[string]any, len(v.Group()))
		for _, sub := range v.Group() {
			m[sub.Key] = sub.Value.Resolve().Any()
		}

		return zap.Any(a.Key, m)
	default:
		return zap.Any(a.Key, v.Any())
	}
}
