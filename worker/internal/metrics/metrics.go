package metrics

import (
	"log/slog"
	"time"
)

type ProcessingMetrics struct {
	Operation string
	Status    string
	Duration  time.Duration
	Size      int64
}

func Record(m ProcessingMetrics) {
	slog.Info("metrics",
		"operation", m.Operation,
		"status", m.Status,
		"duration_ms", m.Duration.Milliseconds(),
		"size", m.Size,
	)
}