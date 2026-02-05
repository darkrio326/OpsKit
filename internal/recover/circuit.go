package recover

import (
	"encoding/json"
	"os"
	"time"

	"opskit/internal/core/fsx"
)

type CircuitState struct {
	LastFailureTime string `json:"lastFailureTime,omitempty"`
	CooldownUntil   string `json:"cooldownUntil,omitempty"`
	LastError       string `json:"lastError,omitempty"`
	LastRunTime     string `json:"lastRunTime,omitempty"`
	LastStatus      string `json:"lastStatus,omitempty"`
	LastTrigger     string `json:"lastTrigger,omitempty"`
	SuccessCount    int    `json:"successCount,omitempty"`
	FailureCount    int    `json:"failureCount,omitempty"`
	WarnCount       int    `json:"warnCount,omitempty"`
}

func Load(path string) (CircuitState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CircuitState{}, nil
		}
		return CircuitState{}, err
	}
	var c CircuitState
	if err := json.Unmarshal(b, &c); err != nil {
		return CircuitState{}, err
	}
	return c, nil
}

func IsOpen(c CircuitState, now time.Time) (bool, time.Time) {
	if c.CooldownUntil == "" {
		return false, time.Time{}
	}
	t, err := time.Parse(time.RFC3339, c.CooldownUntil)
	if err != nil {
		return false, time.Time{}
	}
	return now.Before(t), t
}

func Open(path string, now time.Time, cooldown time.Duration, lastErr string) error {
	return OpenWithTrigger(path, now, cooldown, lastErr, "")
}

func Close(path string) error {
	return CloseWithTrigger(path, time.Now(), "")
}

func OpenWithTrigger(path string, now time.Time, cooldown time.Duration, lastErr string, trigger string) error {
	state, err := Load(path)
	if err != nil {
		return err
	}
	state.LastFailureTime = now.Format(time.RFC3339)
	state.CooldownUntil = now.Add(cooldown).Format(time.RFC3339)
	state.LastError = lastErr
	state.LastRunTime = now.Format(time.RFC3339)
	state.LastStatus = "FAILED"
	state.LastTrigger = trigger
	state.FailureCount++
	return fsx.AtomicWriteJSON(path, state)
}

func CloseWithTrigger(path string, now time.Time, trigger string) error {
	state, err := Load(path)
	if err != nil {
		return err
	}
	state.CooldownUntil = ""
	state.LastError = ""
	state.LastRunTime = now.Format(time.RFC3339)
	state.LastStatus = "PASSED"
	state.LastTrigger = trigger
	state.SuccessCount++
	return fsx.AtomicWriteJSON(path, state)
}

func MarkWarn(path string, now time.Time, reason string, trigger string) error {
	state, err := Load(path)
	if err != nil {
		return err
	}
	state.LastRunTime = now.Format(time.RFC3339)
	state.LastStatus = "WARN"
	state.LastTrigger = trigger
	state.LastError = reason
	state.WarnCount++
	return fsx.AtomicWriteJSON(path, state)
}
