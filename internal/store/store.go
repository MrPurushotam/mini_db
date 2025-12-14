package store

import (
	"sync"

	"github.com/mrpurushotam/mini_database/internal/logger"
)

type Store struct {
	mu     sync.RWMutex
	data   map[string]string
	logger logger.Logger
}

func NewStore(l logger.Logger) *Store {
	return &Store{
		data:   make(map[string]string),
		logger: l,
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	s.logger.Debug("Set operation", "key", key, "Value", value)
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	s.logger.Debug("Get operation", "key", key, "exists", exists)
	return value, exists
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		s.logger.Info("Deleted key", "key", key)
	} else {
		s.logger.Warn("Attempted to delete non-existent key", "key", key)
	}
	return exists
}

func (s *Store) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	s.logger.Debug("GetAll operation", "count", len(result))
	return result
}

func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	s.logger.Debug("GetAllKeys operation", "count", len(keys))
	return keys
}

func (s *Store) GetAllValues() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]string, 0, len(s.data))
	for _, v := range s.data {
		values = append(values, v)
	}
	s.logger.Debug("GetAllValues operation", "count", len(values))
	return values
}
