package dep

import (
	"context"
	"log/slog"
	"os"
	"time"

	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/lmittmann/tint"
)

// GetLogger returns a slog.Logger instance based on the provided log level and environment mode.
// In production mode, it uses Sentry for error logging if a DSN is provided; otherwise, it defaults to JSON logging.
// In non-production modes, it uses a human-friendly console logger.
func GetLogger(level slog.Leveler, appMode AppModeType, sentryDsn string) *slog.Logger {
	var handler slog.Handler

	if appMode == EnvProd {
		if sentryDsn != "" {
			ctx := context.Background()
			handler = sentryslog.Option{
				EventLevel: []slog.Level{slog.LevelError},                // Captures only [slog.LevelError] as error events.
				LogLevel:   []slog.Level{slog.LevelWarn, slog.LevelInfo}, // Captures only [slog.LevelWarn] and [slog.LevelInfo] as log entries.
			}.NewSentryHandler(ctx)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level:     level,
				AddSource: true,
			})
		}
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		})
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
	return logger
}

// LogFatalErr logs the error message and exits the program if err is not nil.
func LogFatalErr(logger *slog.Logger, err error, msg string) {
	if err != nil {
		logger.Error(msg, "err", err)
		os.Exit(1)
	}
}
