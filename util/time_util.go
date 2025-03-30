package util

import (
	"time"
)

func ParseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}

func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}
