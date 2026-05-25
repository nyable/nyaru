package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunHookedCommand checks if 'sfsu' is available and runs it with --json,
// otherwise runs the default scoop command hooked into powershell's ConvertTo-Json.
// Returns: stdout, stderr, cmdStr, isSfsu, err
func RunHookedCommand(action string, query string) (string, string, string, bool, error) {
	_, err := exec.LookPath("sfsu")
	if err == nil {
		// sfsu is available
		var args []string
		// Split action into multiple arguments (e.g. "bucket list" -> ["bucket", "list"])
		args = append(args, strings.Fields(action)...)
		if query != "" {
			args = append(args, query)
		}
		args = append(args, "--json")

		cmd := exec.Command("sfsu", args...)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		cmdStr := strings.Join(cmd.Args, " ")
		err := cmd.Run()
		return stdoutBuf.String(), stderrBuf.String(), cmdStr, true, err
	}

	// fallback to scoop and powershell
	pureCmdStr := fmt.Sprintf("scoop %s", action)
	if query != "" {
		pureCmdStr += " " + query
	}

	psCmdStr := fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr)
	out, sOut, err := RunWithPowerShellCombined("powershell", "-Command", psCmdStr)
	return out, "", sOut, false, err
}
