package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunHookedCommand checks if 'sfsu' is available and runs it with --json, 
// otherwise runs the default scoop command hooked into powershell's ConvertTo-Json.
// Returns: output string, command string, boolean indicates if sfsu was used, error
func RunHookedCommand(action string, query string) (string, string, bool, error) {
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
		cmdStr := strings.Join(cmd.Args, " ")
		output, err := cmd.CombinedOutput()
		return string(output), cmdStr, true, err
	}

	// fallback to scoop and powershell
	pureCmdStr := fmt.Sprintf("scoop %s", action)
	if query != "" {
		pureCmdStr += " " + query
	}
	
	psCmdStr := fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr)
	out, sOut, err := RunWithPowerShellCombined("powershell", "-Command", psCmdStr)
	return out, sOut, false, err
}
