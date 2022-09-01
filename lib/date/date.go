package date

import (
	"fmt"
	"strings"
	"time"
)

func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
}

func StartAndEndOfDay(t time.Time) (time.Time, time.Time) {
	y, m, d := t.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)
	return start, end
}

func SecondsInHumanReadableFormat(n int) string {
	mins := n / 60
	hours := mins / 60
	mins = mins % 60
	days := hours / 24
	hours = hours % 24

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%v days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%v hours", hours))
	}
	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%v mins", mins))
	}

	return strings.Join(parts, " ")
}
