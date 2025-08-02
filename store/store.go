package store

import (
	"path/filepath"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	return value, ok
}

func (s *Store) Match(pattern string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var found []string
	for key := range s.data {
		matched, _ := filepath.Match(pattern, key)
		if matched {
			found = append(found, key)
		}
	}
	if len(found) == 0 {
		return found, false
	}
	return found, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, existed := s.data[key]
	delete(s.data, key)
	return existed
}

func (s *Store) FlushAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]string)
}
