package tui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
)

const maxCapturedOutput = 512 * 1024

// captureBuffer keeps a bounded copy of the command output while the same
// bytes are still written to the terminal. A mutex is required because
// stdout and stderr are pumped by separate goroutines.
type captureBuffer struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func newCaptureBuffer(limit int) *captureBuffer {
	return &captureBuffer{limit: limit}
}

func (c *captureBuffer) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.buf.Len() < c.limit {
		remaining := c.limit - c.buf.Len()
		if len(p) > remaining {
			c.buf.Write(p[:remaining])
			c.truncated = true
		} else {
			c.buf.Write(p)
		}
	} else if len(p) > 0 {
		c.truncated = true
	}
	return len(p), nil
}

func (c *captureBuffer) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := ansi.Strip(c.buf.String())
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\r", "\n")
	out = strings.ReplaceAll(out, "\t", "    ")
	out = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r >= 32 {
			return r
		}
		return -1
	}, out)
	if c.truncated {
		out += "\n...输出过长，已截断..."
	}
	return out
}

// commandWithBanner runs a scoop command after printing a short banner.
// Bubble Tea releases the terminal before Run, so this text and the command's
// own output stay visible in the scrollback while the TUI is paused. Output is
// also captured for the in-app result screen, which remains open until the
// user closes it.
type commandWithBanner struct {
	cmd     *exec.Cmd
	banner  string
	stdout  io.Writer
	capture *captureBuffer
}

func newCommandWithBanner(cmd *exec.Cmd, banner string) *commandWithBanner {
	return &commandWithBanner{
		cmd:     cmd,
		banner:  banner,
		capture: newCaptureBuffer(maxCapturedOutput),
	}
}

func (c *commandWithBanner) SetStdin(r io.Reader) {
	if c.cmd.Stdin == nil {
		c.cmd.Stdin = r
	}
}

func (c *commandWithBanner) SetStdout(w io.Writer) {
	c.stdout = w
	c.cmd.Stdout = io.MultiWriter(w, c.capture)
}

func (c *commandWithBanner) SetStderr(w io.Writer) {
	c.cmd.Stderr = io.MultiWriter(w, c.capture)
}

func (c *commandWithBanner) Run() error {
	out := c.stdout
	if out == nil {
		out = os.Stdout
	}
	if c.banner != "" {
		fmt.Fprintln(out, c.banner)
	}
	return c.cmd.Run()
}

func (c *commandWithBanner) Output() string {
	return c.capture.String()
}

// commandExitCode returns the process exit code, or -1 when the command never
// started (for example, the executable was not found).
func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
