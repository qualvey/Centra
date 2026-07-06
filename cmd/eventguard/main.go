package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"eventguard/internal/action"
	"eventguard/internal/config"
	"eventguard/internal/engine"
	"eventguard/internal/parser/singbox"
	"eventguard/internal/reader"
	"eventguard/internal/rule"
	"eventguard/internal/storage"
)

func main() {
	configPath := flag.String("config", "", "optional JSON config file")
	flag.String("source", "", "log source: stdin or journalctl")
	flag.String("unit", "", "systemd unit used when source=journalctl")
	flag.Int("threshold", 0, "event count threshold per IP")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	applyFlagOverrides(&cfg, mapVisitedFlags())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logReader, closeFn, err := buildReader(ctx, cfg.Source, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if closeFn != nil {
		defer closeFn()
	}

	stateStore, closeStore, err := buildStorage(cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}
	if closeStore != nil {
		defer closeStore()
	}

	rules := []engine.Rule{}
	if cfg.Rules.RealityInvalidHandshake.Enabled {
		rules = append(rules, rule.NewIPThresholdRule(rule.IPThresholdConfig{
			EventType: singbox.RealityInvalidHandshake,
			Threshold: cfg.Rules.RealityInvalidHandshake.Threshold,
		}))
	}

	processor := engine.New(engine.Config{
		Reader:  logReader,
		Parser:  singbox.NewParser(),
		Storage: stateStore,
		Rules:   rules,
		Actions: []engine.Action{
			action.NewConsoleSuggestion(os.Stdout),
		},
	})

	if err := processor.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

func buildStorage(cfg config.StorageConfig) (engine.Storage, func() error, error) {
	switch strings.ToLower(cfg.Type) {
	case "", "memory":
		return storage.NewMemoryStore(), nil, nil
	case "sqlite":
		store, err := storage.OpenSQLite(cfg.Path)
		if err != nil {
			return nil, nil, err
		}
		return store, store.Close, nil
	default:
		return nil, nil, fmt.Errorf("unknown storage type %q", cfg.Type)
	}
}

func mapVisitedFlags() map[string]bool {
	visited := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})
	return visited
}

func applyFlagOverrides(cfg *config.Config, visited map[string]bool) {
	if visited["source"] {
		cfg.Source.Type = flag.Lookup("source").Value.String()
	}
	if visited["unit"] {
		cfg.Source.Unit = flag.Lookup("unit").Value.String()
	}
	if visited["threshold"] {
		value, ok := flag.Lookup("threshold").Value.(flag.Getter).Get().(int)
		if ok && value > 0 {
			cfg.Rules.RealityInvalidHandshake.Threshold = value
		}
	}
}

func buildReader(ctx context.Context, source config.SourceConfig, stdin io.Reader) (engine.Reader, func() error, error) {
	switch strings.ToLower(source.Type) {
	case "stdin":
		return reader.NewScannerReader(bufio.NewScanner(stdin)), nil, nil
	case "journalctl":
		journal, err := reader.NewJournalctlReaderWithConfig(ctx, reader.JournalctlConfig{
			Unit:           source.Unit,
			HistoryEnabled: source.History.Enabled,
			Since:          source.History.Since,
			Follow:         source.History.Follow,
			Resume:         source.History.Resume,
			CheckpointFile: source.History.CheckpointFile,
		})
		if err != nil {
			return nil, nil, err
		}
		return journal, journal.Close, nil
	default:
		return nil, nil, fmt.Errorf("unknown source %q", source.Type)
	}
}
