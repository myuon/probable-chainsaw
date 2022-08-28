package date

import "time"

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
