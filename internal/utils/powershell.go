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
		if strings.HasPrefix(cleanLine, "[") {
			outputJSONStr = cleanLine
			break
		} else if strings.HasPrefix(cleanLine, "{") {
			outputJSONStr = "[" + cleanLine + "]"
			break
		}
	}
	var list []T
	jsonBytes := []byte(outputJSONStr)
	err := json.Unmarshal(jsonBytes, &list)
	return list, err
}
