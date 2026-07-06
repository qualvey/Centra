package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Source SourceConfig `json:"source"`
	Rules  RulesConfig  `json:"rules"`
}

type SourceConfig struct {
	Type string `json:"type"`
	Unit string `json:"unit"`
}

type RulesConfig struct {
	RealityInvalidHandshake IPThresholdConfig `json:"reality_invalid_handshake"`
}

type IPThresholdConfig struct {
	Enabled   bool `json:"enabled"`
	Threshold int  `json:"threshold"`
}

func Default() Config {
	return Config{
		Source: SourceConfig{
			Type: "stdin",
			Unit: "sing-box",
		},
		Rules: RulesConfig{
			RealityInvalidHandshake: IPThresholdConfig{
				Enabled:   true,
				Threshold: 5,
			},
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	normalize(&cfg)
	return cfg, nil
}

func normalize(cfg *Config) {
	if cfg.Source.Type == "" {
		cfg.Source.Type = "stdin"
	}
	if cfg.Source.Unit == "" {
		cfg.Source.Unit = "sing-box"
	}
	if cfg.Rules.RealityInvalidHandshake.Threshold <= 0 {
		cfg.Rules.RealityInvalidHandshake.Threshold = 5
	}
}
