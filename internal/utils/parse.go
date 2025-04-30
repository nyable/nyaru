package utils

import "strings"

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
