package storage

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu     sync.Mutex
	counts map[string]int
	marks  map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		counts: make(map[string]int),
		marks:  make(map[string]struct{}),
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

func (s *MemoryStore) MarkOnce(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.marks[key]; exists {
		return false, nil
	}
	s.marks[key] = struct{}{}
	return true, nil
}
