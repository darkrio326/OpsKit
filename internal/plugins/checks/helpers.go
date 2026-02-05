package checks

import (
	"fmt"
	"strconv"
	"strings"

	"opskit/internal/schema"
)

func toStringSlice(v any, fallback []string) []string {
	arr, ok := v.([]any)
	if !ok {
		if s, ok2 := v.([]string); ok2 {
			return s
		}
		return fallback
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		out = append(out, fmt.Sprintf("%v", item))
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func toIntSlice(v any, fallback []int) []int {
	arr, ok := v.([]any)
	if !ok {
		if ints, ok2 := v.([]int); ok2 {
			return ints
		}
		if stringsArr, ok2 := v.([]string); ok2 {
			out := make([]int, 0, len(stringsArr))
			for _, s := range stringsArr {
				if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
					out = append(out, n)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
		return fallback
	}
	out := make([]int, 0, len(arr))
	for _, item := range arr {
		switch n := item.(type) {
		case float64:
			out = append(out, int(n))
		case int:
			out = append(out, n)
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
				out = append(out, parsed)
			}
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func toInt(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return parsed
		}
		return fallback
	default:
		return fallback
	}
}

func toString(v any, fallback string) string {
	if v == nil {
		return fallback
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return fallback
	}
	return s
}

func toBool(v any, fallback bool) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func toSeverity(v any, fallback schema.Severity) schema.Severity {
	switch strings.ToLower(toString(v, "")) {
	case "info":
		return schema.SeverityInfo
	case "warn":
		return schema.SeverityWarn
	case "fail":
		return schema.SeverityFail
	default:
		return fallback
	}
}

func statusFromSeverity(sev schema.Severity) schema.Status {
	switch sev {
	case schema.SeverityWarn:
		return schema.StatusWarn
	case schema.SeverityInfo:
		return schema.StatusPassed
	default:
		return schema.StatusFailed
	}
}
