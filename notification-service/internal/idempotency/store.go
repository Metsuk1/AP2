package idempotency

import "sync"

type Store struct {
	mu        sync.Mutex
	processed map[string]struct{}
}

func NewStore() *Store {
	return &Store{processed: make(map[string]struct{})}
}

func (s *Store) IsProcessed(eventID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.processed[eventID]
	return ok
}

func (s *Store) MarkProcessed(eventID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processed[eventID] = struct{}{}
}
