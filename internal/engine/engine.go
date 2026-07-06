package engine

import (
	"context"
	"errors"
	"io"

	"eventguard/internal/core"
)

type Config struct {
	Reader  Reader
	Parser  Parser
	Storage Storage
	Rules   []Rule
	Actions []Action
}

type Engine struct {
	config Config
}

func New(config Config) *Engine {
	return &Engine{config: config}
}

func (e *Engine) Run(ctx context.Context) error {
	for {
		line, err := e.config.Reader.ReadLine(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		event, ok, err := e.config.Parser.Parse(line)
		if err != nil {
			return err
		}
		if !ok {
			if err := e.commitLine(ctx, line); err != nil {
				return err
			}
			continue
		}

		if err := e.saveEvent(ctx, event); err != nil {
			return err
		}

		for _, r := range e.config.Rules {
			triggers, err := r.Evaluate(ctx, event, e.config.Storage)
			if err != nil {
				return err
			}
			for _, trigger := range triggers {
				for _, act := range e.config.Actions {
					if err := act.Execute(ctx, trigger); err != nil {
						return err
					}
				}
			}
		}

		if err := e.commitLine(ctx, line); err != nil {
			return err
		}
	}
}

func (e *Engine) commitLine(ctx context.Context, line string) error {
	committer, ok := e.config.Reader.(LineCommitter)
	if !ok {
		return nil
	}
	return committer.CommitLine(ctx, line)
}

func (e *Engine) saveEvent(ctx context.Context, event core.Event) error {
	recorder, ok := e.config.Storage.(EventRecorder)
	if !ok {
		return nil
	}
	return recorder.SaveEvent(ctx, event)
}
