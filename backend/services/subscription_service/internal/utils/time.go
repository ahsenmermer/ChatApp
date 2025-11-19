package utils

import "time"

// AddDays returns a new time after adding given days to now (UTC)
func AddDays(days int) time.Time {
	return time.Now().UTC().AddDate(0, 0, days)
}
