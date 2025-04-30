package utils

import (
	"fmt"
	"math"
	"strings"
)

func GetSource(source string) string {
	if strings.HasPrefix(source, "http") {
		lastSlashIndex := strings.LastIndex(source, "/")
		if lastSlashIndex != -1 && lastSlashIndex < len(source)-1 {
			// Return the substring after the last slash
			return source[lastSlashIndex+1:]
		}
	}
	return source
}

func SizeToHumanRead(size int64) string {
	// Handle base case: bytes less than 1024 or non-positive
	if size <= 0 {
		return "0 B"
	}

	const base = 1024 // Use 1024 for KB, MB, GB calculation
	// Define the units
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"} // Exabytes should be sufficient for most cases

	// Calculate the appropriate unit index
	i := math.Floor(math.Log(float64(size)) / math.Log(base))

	// Ensure the index is within the bounds of the units slice
	if i >= float64(len(units)) {
		i = float64(len(units) - 1) // Cap at the largest unit
	}

	// Calculate the value in the determined unit
	value := float64(size) / math.Pow(base, i)

	// Handle the case where the value should be displayed as Bytes (i=0)
	if i == 0 {
		return fmt.Sprintf("%d %s", size, units[int(i)])
	}

	// Format the value with one decimal place and the unit
	return fmt.Sprintf("%.1f %s", value, units[int(i)])
}
