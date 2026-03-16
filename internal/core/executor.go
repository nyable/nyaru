package core

import (
	"bytes"
	"os"
	"os/exec"
)

// runInteractiveCommand runs a command where its output is sent to stdout/stderr
func runInteractiveCommand(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// runCommandOutput runs a command and returns its combined stdout and stderr
func runCommandOutput(cmdName string, args ...string) (string, error) {
	cmd := exec.Command(cmdName, args...)
	var outBytes bytes.Buffer
	cmd.Stdout = &outBytes
	cmd.Stderr = &outBytes

	err := cmd.Run()
	return outBytes.String(), err
}
