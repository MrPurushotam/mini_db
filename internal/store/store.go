package store

import (
	"github.com/mrpurushotam/mini_database/internal/aof"
	"github.com/mrpurushotam/mini_database/internal/logger"
	"sync"
)

type Store struct {
	mu        sync.RWMutex
	data      map[string]string
	aof       *aof.AOF
	enableAof bool
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) EnableAOF(aofInstance *aof.AOF) {
	s.aof = aofInstance
	if aofInstance != nil {
		s.enableAof = true
	}
}

func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value

	if s.aof != nil {
		if err := s.aof.Write("SET", key, value); err != nil {
			return err
		}
	}
	logger.Debug("Set operation", "key", key, "Value", value)
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	logger.Debug("Get operation", "key", key, "exists", exists)
	return value, exists
}

func (s *Store) Delete(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		logger.Info("Deleted key", "key", key)

		if s.aof != nil {
			if err := s.aof.Write("DELETE", key, ""); err != nil {
				return false, err
			}
		}

	} else {
		logger.Warn("Attempted to delete non-existent key", "key", key)
	}
	return exists, nil
}

func (s *Store) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	logger.Debug("GetAll operation", "count", len(result))
	return result
}

func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	logger.Debug("GetAllKeys operation", "count", len(keys))
	return keys
}

func (s *Store) GetAllValues() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]string, 0, len(s.data))
	for _, v := range s.data {
		values = append(values, v)
	}
	logger.Debug("GetAllValues operation", "count", len(values))
	return values
}

func (s *Store) LoadFromAOF(filepath string) error {
	tempAOF := &aof.AOF{}
	logger.Info("Loading data from AOF...")

	operations, err := tempAOF.Read(filepath)
	if err != nil {
		return err
	}

	for _, op := range operations {
		switch op.Type {
		case "SET":
			s.data[op.Key] = op.Value
		case "DELETE":
			delete(s.data, op.Key)
		}
	}
	logger.Info("AOF loaded successfully")
	return nil
}
