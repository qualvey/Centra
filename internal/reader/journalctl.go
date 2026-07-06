package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type JournalctlReader struct {
	cmd     *exec.Cmd
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

func NewJournalctlReader(ctx context.Context, unit string) (*JournalctlReader, error) {
	args := []string{"-f", "-o", "short-iso"}
	if unit != "" {
		args = append(args, "-u", unit)
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start journalctl: %w", err)
	}

	return &JournalctlReader{
		cmd:     cmd,
		stdout:  stdout,
		scanner: bufio.NewScanner(stdout),
	}, nil
}

func (r *JournalctlReader) ReadLine(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}

func (r *JournalctlReader) Close() error {
	if r.stdout != nil {
		_ = r.stdout.Close()
	}
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}
	_ = r.cmd.Process.Kill()
	return r.cmd.Wait()
}
