package storage

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		counts: make(map[string]int),
	}
}

func (s *MemoryStore) Increment(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[key]++
	return s.counts[key], nil
}
