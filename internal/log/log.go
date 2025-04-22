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

var defaultLoggerOptions = &slog.HandlerOptions{
	Level: Level,
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			levelLabel, exists := levelNames[level]
			if !exists {
				levelLabel = level.String()
			}

			a.Value = slog.StringValue(levelLabel)
		}

		return a
	},
}

var Level = &slog.LevelVar{}

var Logger = slog.New(slog.NewTextHandler(os.Stdout, defaultLoggerOptions))

// Initialise default logger.
func init() {
	slog.SetDefault(Logger)
}
