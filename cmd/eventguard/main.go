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
	"eventguard/internal/engine"
	"eventguard/internal/parser/singbox"
	"eventguard/internal/reader"
	"eventguard/internal/rule"
	"eventguard/internal/storage"
)

func main() {
	var (
		source    = flag.String("source", "stdin", "log source: stdin or journalctl")
		unit      = flag.String("unit", "sing-box", "systemd unit used when source=journalctl")
		threshold = flag.Int("threshold", 5, "event count threshold per IP")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logReader, closeFn, err := buildReader(ctx, *source, *unit, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if closeFn != nil {
		defer closeFn()
	}

	processor := engine.New(engine.Config{
		Reader:  logReader,
		Parser:  singbox.NewParser(),
		Storage: storage.NewMemoryStore(),
		Rules: []engine.Rule{
			rule.NewIPThresholdRule(rule.IPThresholdConfig{
				EventType: "singbox.reality_invalid_handshake",
				Threshold: *threshold,
			}),
		},
		Actions: []engine.Action{
			action.NewConsoleSuggestion(os.Stdout),
		},
	})

	if err := processor.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

func buildReader(ctx context.Context, source, unit string, stdin io.Reader) (engine.Reader, func() error, error) {
	switch strings.ToLower(source) {
	case "stdin":
		return reader.NewScannerReader(bufio.NewScanner(stdin)), nil, nil
	case "journalctl":
		journal, err := reader.NewJournalctlReader(ctx, unit)
		if err != nil {
			return nil, nil, err
		}
		return journal, journal.Close, nil
	default:
		return nil, nil, fmt.Errorf("unknown source %q", source)
	}
}
