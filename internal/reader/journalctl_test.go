package reader

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildJournalctlArgsRealtime(t *testing.T) {
	args, err := BuildJournalctlArgs(JournalctlConfig{
		Unit:   "sing-box",
		Follow: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"-o", "short-iso", "-u", "sing-box", "-f"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestBuildJournalctlArgsHistoryWithFollow(t *testing.T) {
	args, err := BuildJournalctlArgs(JournalctlConfig{
		Unit:           "sing-box",
		HistoryEnabled: true,
		Since:          "2026-07-06 00:00:00",
		Follow:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"-o", "short-iso", "--no-pager", "-u", "sing-box", "--since", "2026-07-06 00:00:00", "-f", "--no-tail"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestBuildJournalctlArgsUsesCheckpointBeforeSince(t *testing.T) {
	checkpoint := filepath.Join(t.TempDir(), "checkpoint")
	if err := os.WriteFile(checkpoint, []byte("2026-07-06T21:47:27+08:00\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	args, err := BuildJournalctlArgs(JournalctlConfig{
		Unit:           "sing-box",
		HistoryEnabled: true,
		Since:          "2026-07-06 00:00:00",
		Follow:         false,
		Resume:         true,
		CheckpointFile: checkpoint,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"-o", "short-iso", "--no-pager", "-u", "sing-box", "--since", "2026-07-06T21:47:27+08:00"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestCommitLineWritesCheckpoint(t *testing.T) {
	checkpoint := filepath.Join(t.TempDir(), "state", "checkpoint")
	reader := &JournalctlReader{checkpointFile: checkpoint}

	err := reader.CommitLine(context.Background(), "2026-07-06T21:47:27+08:00 host sing-box[1]: message")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "2026-07-06T21:47:27+08:00\n" {
		t.Fatalf("checkpoint = %q", string(data))
	}
}
