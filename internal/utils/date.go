package utils

import (
	"strconv"
	"strings"
	"time"
)

/*
* Format date string with prefix and suffix
 */
func FormatDateWithWrapper(dateStr string, prefix string, suffix string, pattern string) string {
	if !strings.HasPrefix(dateStr, prefix) || !strings.HasSuffix(dateStr, suffix) {
		return "Error Wrapper!"
	}
	numStr := dateStr[len(prefix) : len(dateStr)-len(suffix)]
	timestampMillis, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return "Error Value!"
	}
	timestampSeconds := timestampMillis / 1000
	t := time.Unix(timestampSeconds, 0)
	if pattern == "" {
		pattern = "2006-01-02 15:04:05"
	}
	formattedTime := t.Local().Format(pattern)
	return formattedTime
}
