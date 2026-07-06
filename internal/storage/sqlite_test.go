package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"eventguard/internal/core"
)

func TestSQLiteStorePersistsEventsAndIPState(t *testing.T) {
	store := openTestSQLite(t)
	defer store.Close()

	event := core.Event{
		Timestamp: time.Date(2026, 7, 6, 21, 47, 27, 0, time.UTC),
		Source:    "journalctl",
		Service:   "sing-box",
		EventType: "singbox.reality_invalid_handshake",
		Level:     "error",
		IP:        "45.194.67.28",
		Message:   "test message",
		Metadata: map[string]string{
			"parser": "singbox",
		},
	}

	if err := store.SaveEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}

	count, err := store.CountEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("event count = %d, want 1", count)
	}

	state, ok, err := store.GetIPState(context.Background(), "45.194.67.28")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ip state")
	}
	if state.Count != 1 {
		t.Fatalf("ip state count = %d, want 1", state.Count)
	}
	if state.Status != "observed" {
		t.Fatalf("ip state status = %q", state.Status)
	}
}

func TestSQLiteStoreIncrementAndMarkOnce(t *testing.T) {
	store := openTestSQLite(t)
	defer store.Close()

	value, err := store.Increment(context.Background(), "event_count:test:1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if value != 1 {
		t.Fatalf("first counter value = %d", value)
	}

	value, err = store.Increment(context.Background(), "event_count:test:1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if value != 2 {
		t.Fatalf("second counter value = %d", value)
	}

	first, err := store.MarkOnce(context.Background(), "triggered:test:1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("first mark should return true")
	}

	second, err := store.MarkOnce(context.Background(), "triggered:test:1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if second {
		t.Fatal("second mark should return false")
	}
}

func openTestSQLite(t *testing.T) *SQLiteStore {
	t.Helper()

	store, err := OpenSQLite(filepath.Join(t.TempDir(), "eventguard.db"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}
