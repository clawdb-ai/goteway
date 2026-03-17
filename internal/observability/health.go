package observability

import "time"

// HealthReporter provides basic liveness/readiness payloads.
type HealthReporter struct {
	startedAt time.Time
	version   string
}

func NewHealthReporter(version string) *HealthReporter {
	return &HealthReporter{startedAt: time.Now(), version: version}
}

func (h *HealthReporter) Payload() map[string]any {
	return map[string]any{
		"status":    "ok",
		"version":   h.version,
		"uptimeSec": int(time.Since(h.startedAt).Seconds()),
		"deps": map[string]any{
			"store":         "ok",
			"pluginRuntime": "ok",
		},
	}
}
