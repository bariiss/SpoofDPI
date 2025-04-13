package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bariiss/SpoofDPI/util"
	"github.com/rs/zerolog"
)

const (
	scopeFieldName   = "scope"
	traceIdFieldName = "trace_id"
)

var logger zerolog.Logger

// GetCtxLogger returns a logger with the context's trace ID and scope.
func GetCtxLogger(ctx context.Context) zerolog.Logger {
	return logger.With().Ctx(ctx).Logger()
}

// InitLogger initializes the logger with the given configuration.
func InitLogger(cfg *util.Config) {
	partsOrder := []string{
		zerolog.LevelFieldName,
		zerolog.TimestampFieldName,
		traceIdFieldName,
		scopeFieldName,
		zerolog.MessageFieldName,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    time.RFC3339,
		PartsOrder:    partsOrder,
		FieldsExclude: []string{traceIdFieldName, scopeFieldName},
		FormatPrepare: func(m map[string]any) error {
			formatFieldValue(m, "%s", traceIdFieldName)
			formatFieldValue(m, "[%s]", scopeFieldName)
			return nil
		},
	}

	level := zerolog.InfoLevel
	if cfg.Debug {
		level = zerolog.DebugLevel
	}

	logger = zerolog.New(consoleWriter).
		Level(level).
		Hook(ctxHook{}).
		With().
		Timestamp().
		Logger()
}

// formatFieldValue formats the field value in the map according to the given format.
func formatFieldValue(vs map[string]any, format string, field string) {
	value, ok := vs[field]
	if !ok {
		vs[field] = ""
		return
	}

	switch v := value.(type) {
	case string:
		vs[field] = fmt.Sprintf(format, v)
	default:
		vs[field] = ""
	}
}

type ctxHook struct{}

// Run adds the trace ID and scope to the log event from the context.
func (h ctxHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	if scope, ok := util.GetScopeFromCtx(ctx); ok {
		e.Str(scopeFieldName, scope)
	}

	if traceId, ok := util.GetTraceIdFromCtx(ctx); ok {
		e.Str(traceIdFieldName, traceId)
	}
}
