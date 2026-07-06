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
	if !cfg.Rules.RealityInvalidHandshake.Enabled {
		t.Fatal("reality invalid handshake rule should be enabled by default")
	}
	if cfg.Rules.RealityInvalidHandshake.Threshold != 5 {
		t.Fatalf("threshold = %d", cfg.Rules.RealityInvalidHandshake.Threshold)
	}
}

func TestLoadConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := os.WriteFile(path, []byte(`{
		"source": {"type": "journalctl", "unit": "sing-box.service"},
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
	if cfg.Rules.RealityInvalidHandshake.Threshold != 7 {
		t.Fatalf("threshold = %d", cfg.Rules.RealityInvalidHandshake.Threshold)
	}
}
