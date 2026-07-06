package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Source.Type != "stdin" {
		t.Fatalf("source type = %q", cfg.Source.Type)
	}
	if cfg.Source.Unit != "sing-box" {
		t.Fatalf("source unit = %q", cfg.Source.Unit)
	}
	if cfg.Storage.Type != "memory" {
		t.Fatalf("storage type = %q", cfg.Storage.Type)
	}
	if cfg.Storage.Path != ".eventguard/eventguard.db" {
		t.Fatalf("storage path = %q", cfg.Storage.Path)
	}
	if !cfg.Source.History.Follow {
		t.Fatal("journal history follow should default to true")
	}
	if !cfg.Rules.RealityInvalidHandshake.Enabled {
		t.Fatal("reality invalid handshake rule should be enabled by default")
	}
	if cfg.Rules.RealityInvalidHandshake.Threshold != 5 {
		t.Fatalf("threshold = %d", cfg.Rules.RealityInvalidHandshake.Threshold)
	}
}

func TestLoadHistoryConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := os.WriteFile(path, []byte(`{
		"source": {
			"type": "journalctl",
			"unit": "sing-box",
			"history": {
				"enabled": true,
				"since": "2026-07-06 00:00:00",
				"follow": true,
				"resume": true
			}
		}
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Source.History.Enabled {
		t.Fatal("history should be enabled")
	}
	if cfg.Source.History.Since != "2026-07-06 00:00:00" {
		t.Fatalf("history since = %q", cfg.Source.History.Since)
	}
	if !cfg.Source.History.Resume {
		t.Fatal("history resume should be enabled")
	}
	if cfg.Source.History.CheckpointFile != ".eventguard/journalctl.checkpoint" {
		t.Fatalf("checkpoint file = %q", cfg.Source.History.CheckpointFile)
	}
}

func TestLoadConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := os.WriteFile(path, []byte(`{
		"source": {"type": "journalctl", "unit": "sing-box.service"},
		"storage": {"type": "sqlite", "path": "/tmp/eventguard.db"},
		"rules": {
			"reality_invalid_handshake": {"enabled": true, "threshold": 7}
		}
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Source.Type != "journalctl" {
		t.Fatalf("source type = %q", cfg.Source.Type)
	}
	if cfg.Source.Unit != "sing-box.service" {
		t.Fatalf("source unit = %q", cfg.Source.Unit)
	}
	if cfg.Storage.Type != "sqlite" {
		t.Fatalf("storage type = %q", cfg.Storage.Type)
	}
	if cfg.Storage.Path != "/tmp/eventguard.db" {
		t.Fatalf("storage path = %q", cfg.Storage.Path)
	}
	if cfg.Rules.RealityInvalidHandshake.Threshold != 7 {
		t.Fatalf("threshold = %d", cfg.Rules.RealityInvalidHandshake.Threshold)
	}
}
