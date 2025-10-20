package item

import (
	"log/slog"
	"time"
)

type Level slog.Level

const (
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelError Level = Level(slog.LevelError)
)

type Item struct {
	Source  string
	Time    time.Time
	Level   Level
	Message string
	Attrs   map[string]any
}
