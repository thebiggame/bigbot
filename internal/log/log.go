package log

import (
	"log/slog"
	"os"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

var Levels = map[string]slog.Level{
	"TRACE": LevelTrace,
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
	"FATAL": LevelFatal,
}

var defaultLoggerOptions = &slog.HandlerOptions{
	Level:       Level,
	ReplaceAttr: Replacer,
}

var Level = &slog.LevelVar{}

var Logger = slog.New(slog.NewTextHandler(os.Stdout, defaultLoggerOptions))

func Replacer(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelLabel, exists := levelNames[level]
		if !exists {
			levelLabel = level.String()
		}

		a.Value = slog.StringValue(levelLabel)
	}

	return a
}

// Initialise default logger.
func init() {
	slog.SetDefault(Logger)
}
