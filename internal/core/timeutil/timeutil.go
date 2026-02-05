package timeutil

import "time"

func NowISO8601() string {
	return time.Now().Format(time.RFC3339)
}

func NowISO8601Compact() string {
	return time.Now().Format("20060102-150405")
}
