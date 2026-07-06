package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var journalTimestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)

type JournalctlConfig struct {
	Unit           string
	HistoryEnabled bool
	Since          string
	Follow         bool
	Resume         bool
	CheckpointFile string
}

type JournalctlReader struct {
	cmd            *exec.Cmd
	stdout         io.ReadCloser
	scanner        *bufio.Scanner
	checkpointFile string
}

func NewJournalctlReader(ctx context.Context, unit string) (*JournalctlReader, error) {
	return NewJournalctlReaderWithConfig(ctx, JournalctlConfig{
		Unit:   unit,
		Follow: true,
	})
}

func NewJournalctlReaderWithConfig(ctx context.Context, config JournalctlConfig) (*JournalctlReader, error) {
	normalizeJournalctlConfig(&config)
	args, err := BuildJournalctlArgs(config)
	if err != nil {
		return nil, err
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
		cmd:            cmd,
		stdout:         stdout,
		scanner:        bufio.NewScanner(stdout),
		checkpointFile: config.CheckpointFile,
	}, nil
}

func BuildJournalctlArgs(config JournalctlConfig) ([]string, error) {
	normalizeJournalctlConfig(&config)

	args := []string{"-o", "short-iso"}
	if config.HistoryEnabled {
		args = append(args, "--no-pager")
	}
	if config.Unit != "" {
		args = append(args, "-u", config.Unit)
	}

	since, err := effectiveSince(config)
	if err != nil {
		return nil, err
	}
	if since != "" {
		args = append(args, "--since", since)
	}

	if config.HistoryEnabled {
		if config.Follow {
			args = append(args, "-f", "--no-tail")
		}
		return args, nil
	}

	args = append(args, "-f")
	return args, nil
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

func (r *JournalctlReader) CommitLine(ctx context.Context, line string) error {
	if r.checkpointFile == "" {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	timestamp := extractJournalTimestamp(line)
	if timestamp == "" {
		return nil
	}
	return writeCheckpoint(r.checkpointFile, timestamp)
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

func normalizeJournalctlConfig(config *JournalctlConfig) {
	if !config.HistoryEnabled {
		config.Follow = true
		config.CheckpointFile = ""
		return
	}
	if config.Resume && config.CheckpointFile == "" {
		config.CheckpointFile = ".eventguard/journalctl.checkpoint"
	}
}

func effectiveSince(config JournalctlConfig) (string, error) {
	if config.HistoryEnabled && config.Resume && config.CheckpointFile != "" {
		checkpoint, err := readCheckpoint(config.CheckpointFile)
		if err != nil {
			return "", err
		}
		if checkpoint != "" {
			return checkpoint, nil
		}
	}
	return config.Since, nil
}

func extractJournalTimestamp(line string) string {
	return journalTimestampPattern.FindString(line)
}

func readCheckpoint(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return "", fmt.Errorf("read journal checkpoint: %w", err)
}

func writeCheckpoint(path, timestamp string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create checkpoint dir: %w", err)
		}
	}

	tmp, err := os.CreateTemp(dir, ".journalctl.checkpoint.*")
	if err != nil {
		return fmt.Errorf("create checkpoint temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(timestamp + "\n"); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write checkpoint: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close checkpoint: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace checkpoint: %w", err)
	}
	return nil
}
