package utils

import (
	"encoding/json"
	"os/exec"
	"strings"
)

func RunWithPowerShellCombined(name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)
	cmdStr := strings.Join(cmd.Args, " ")
	output, err := cmd.CombinedOutput()
	return string(output), cmdStr, err
}

func PsDirtyJSONToStructList[T any](dirtyJSON string) ([]T, error) {
	lines := strings.Split(strings.TrimSpace(dirtyJSON), "\n")
	outputJSONStr := "[]"
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if idx := strings.Index(cleanLine, "["); idx != -1 {
			outputJSONStr = cleanLine[idx:]
			break
		} else if idx := strings.Index(cleanLine, "{"); idx != -1 {
			outputJSONStr = "[" + cleanLine[idx:] + "]"
			break
		}
	}
	var list []T
	jsonBytes := []byte(outputJSONStr)
	err := json.Unmarshal(jsonBytes, &list)
	return list, err
}

func StandardJSONToStructList[T any](jsonStr string) ([]T, error) {
	var list []T
	if strings.TrimSpace(jsonStr) == "" {
		return list, nil
	}
	err := json.Unmarshal([]byte(jsonStr), &list)
	return list, err
}

