package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mrpurushotam/mini_db/internal/aof"
	"github.com/mrpurushotam/mini_db/internal/domain"
	"github.com/mrpurushotam/mini_db/internal/logger"
	DataTypeValue "github.com/mrpurushotam/mini_db/internal/value"
)

type Store struct {
	mu        sync.RWMutex
	data      map[string]domain.Value
	aof       *aof.AOF
	enableAof bool
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]domain.Value),
	}
}

func (s *Store) EnableAOF(aofInstance *aof.AOF) {
	s.aof = aofInstance
	if aofInstance != nil {
		s.enableAof = true
	} else {
		s.enableAof = false
	}
}

func (s *Store) checkType(key string, expectedType domain.DataType) (domain.Value, error) {
	val, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	if val.Type() != expectedType {
		return nil, fmt.Errorf("wrong type: expected %s, got %s", expectedType, val.Type())
	}
	return val, nil
}

// -- String Operations --
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stringValue := &DataTypeValue.StringValue{Data: value}
	s.data[key] = stringValue

	if s.enableAof {
		serialized := stringValue.Serialize()
		if err := s.aof.Write("SET", key, "string", string(serialized)); err != nil {
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
	if !exists {
		return "", false
	}
	stringValue, ok := value.(*DataTypeValue.StringValue)
	if !ok {
		logger.Warn("Type mismatch: expected string", "key", key)
		return "", false
	}
	logger.Debug("Get operation", "key", key, "exists", exists)
	return stringValue.Data, exists
}

// -- Set Operations --

func (s *Store) SAdd(key string, members ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var setVal *DataTypeValue.SetValue

	if !exists {
		setVal = &DataTypeValue.SetValue{Data: make(map[string]struct{})}
		s.data[key] = setVal
	} else {
		var ok bool
		setVal, ok = val.(*DataTypeValue.SetValue)
		if !ok {
			return fmt.Errorf("wrong type: expected set")
		}
	}

	for _, member := range members {
		setVal.Data[member] = struct{}{}
	}

	if s.enableAof {
		for _, member := range members {
			if err := s.aof.Write("SADD", key, "set", member); err != nil {
				return err
			}
		}
	}
	logger.Debug("SADD operation", "key", key, "valueType", "set", "members", members)
	return nil
}

func (s *Store) SMembers(key string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.checkType(key, domain.Set)
	if err != nil {
		return nil, err
	}

	setVal := val.(*DataTypeValue.SetValue)
	members := make([]string, 0, len(setVal.Data))

	for member := range setVal.Data {
		members = append(members, member)
	}

	return members, nil
}

func (s *Store) SPop(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.checkType(key, domain.Set)
	if err != nil {
		return 0, err
	}

	setVal := val.(*DataTypeValue.SetValue)
	removed := 0

	for _, member := range members {
		if _, exists := setVal.Data[member]; exists {
			delete(setVal.Data, member)
			removed++
		}
	}

	if s.enableAof && removed > 0 {
		for _, member := range members {
			if err := s.aof.Write("SPOP", key, "set", member); err != nil {
				return removed, err
			}
		}
	}
	return removed, nil
}

// -- List Operations --

func (s *Store) LPush(key string, values ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var listVal *DataTypeValue.ListValue

	if !exists {
		listVal = &DataTypeValue.ListValue{Data: make([]string, 0)}
		s.data[key] = listVal
	} else {
		var ok bool
		listVal, ok = val.(*DataTypeValue.ListValue)
		if !ok {
			return fmt.Errorf("wrong type: expected list")
		}
	}
	listVal.Data = append(values, listVal.Data...)

	if s.enableAof {
		// LPUSH AOF command should write each value pushed
		for _, v := range values {
			if err := s.aof.Write("LPUSH", key, "list", v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) RPush(key string, values ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var listVal *DataTypeValue.ListValue

	if !exists {
		listVal = &DataTypeValue.ListValue{Data: make([]string, 0)}
		s.data[key] = listVal
	} else {
		var ok bool
		listVal, ok = val.(*DataTypeValue.ListValue)

		if !ok {
			return fmt.Errorf("wrong type: expected list")
		}
	}
	listVal.Data = append(listVal.Data, values...)

	if s.enableAof {
		for _, v := range values {
			if err := s.aof.Write("RPUSH", key, "list", v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.checkType(key, domain.List)
	if err != nil {
		return nil, err
	}

	listVal := val.(*DataTypeValue.ListValue)
	length := len(listVal.Data)

	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}, nil
	}

	return listVal.Data[start : stop+1], nil
}

//--Queue Operations--

func (s *Store) Enqueue(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var queueVal *DataTypeValue.QueueValue

	if !exists {
		queueVal = &DataTypeValue.QueueValue{Data: make([]string, 0)}
		s.data[key] = queueVal
	} else {
		var ok bool
		queueVal, ok = val.(*DataTypeValue.QueueValue)
		if !ok {
			return fmt.Errorf("wrong type: expected queue")
		}
	}

	queueVal.Data = append(queueVal.Data, value)

	if s.enableAof {
		if err := s.aof.Write("ENQUEUE", key, "queue", value); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Dequeue(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.checkType(key, domain.Queue)
	if err != nil {
		return "", err
	}

	queueVal := val.(*DataTypeValue.QueueValue)
	if len(queueVal.Data) == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	value := queueVal.Data[0]
	queueVal.Data = queueVal.Data[1:]

	if s.enableAof {
		// DEQUEUE AOF command should only record the operation, not the dequeued value
		if err := s.aof.Write("DEQUEUE", key, "queue", ""); err != nil {
			return value, err
		}
	}

	return value, nil
}

//-- Stack Operations ---

func (s *Store) Push(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var stackVal *DataTypeValue.StackValue

	if !exists {
		stackVal = &DataTypeValue.StackValue{Data: make([]string, 0)}
		s.data[key] = stackVal
	} else {
		var ok bool
		stackVal, ok = val.(*DataTypeValue.StackValue)
		if !ok {
			return fmt.Errorf("wrong type: expected stack")
		}
	}

	stackVal.Data = append(stackVal.Data, value)
	if s.enableAof {
		if err := s.aof.Write("PUSH", key, "stack", value); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Pop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.checkType(key, domain.Stack)
	if err != nil {
		return "", err
	}

	stackVal := val.(*DataTypeValue.StackValue)
	if len(stackVal.Data) == 0 {
		return "", fmt.Errorf("stack is empty")
	}

	lastIdx := len(stackVal.Data) - 1
	value := stackVal.Data[lastIdx]
	stackVal.Data = stackVal.Data[:lastIdx]

	if s.enableAof {
		// POP AOF command should only record the operation, not the popped value
		if err := s.aof.Write("POP", key, "stack", ""); err != nil {
			return value, err
		}
	}

	return value, nil
}

// ===== HASHMAP OPERATIONS =====

type HSetPayload struct {
	Field string `json:"f"`
	Value string `json:"v"`
}

func (s *Store) HSet(key, field, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	var hashVal *DataTypeValue.HashmapValue

	if !exists {
		hashVal = &DataTypeValue.HashmapValue{Data: make(map[string]string)}
		s.data[key] = hashVal
	} else {
		var ok bool
		hashVal, ok = val.(*DataTypeValue.HashmapValue)
		if !ok {
			return fmt.Errorf("wrong type: expected hashmap")
		}
	}

	hashVal.Data[field] = value

	if s.enableAof {
		payload := HSetPayload{
			Field: field,
			Value: value,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		if err := s.aof.Write("HSET", key, "hashmap", string(data)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) HGet(key, field string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, err := s.checkType(key, domain.Hashmap)
	if err != nil {
		return "", err
	}

	hashVal := val.(*DataTypeValue.HashmapValue)
	value, exists := hashVal.Data[field]
	if !exists {
		return "", fmt.Errorf("field not found")
	}
	return value, nil
}

func (s *Store) HGetAll(key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, err := s.checkType(key, domain.Hashmap)
	if err != nil {
		return nil, err
	}

	hashVal := val.(*DataTypeValue.HashmapValue)
	result := make(map[string]string, len(hashVal.Data))
	for k, v := range hashVal.Data {
		result[k] = v
	}
	return result, nil
}

func (s *Store) Delete(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		logger.Info("Deleted key", "key", key)

		if s.enableAof {
			if err := s.aof.Write("DELETE", key, "", ""); err != nil {
				return false, err
			}
		}

	} else {
		logger.Warn("Attempted to delete non-existent key", "key", key)
	}
	return exists, nil
}

func (s *Store) GetAll() map[string]domain.Value {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]domain.Value, len(s.data))
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

func (s *Store) GetAllValues() []domain.Value {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]domain.Value, 0, len(s.data))
	for _, v := range s.data {
		values = append(values, v)
	}
	logger.Debug("GetAllValues operation", "count", len(values))
	return values
}

func (s *Store) LoadFromAOF(filepath string) error {
	if !s.enableAof {
		logger.Warn("AOF is disabled")
		return nil
	}
	tempAOF := &aof.AOF{}
	logger.Info("Loading data from AOF...")

	operations, err := tempAOF.Read(filepath)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, op := range operations {
		switch op.Type {

		case "SET":
			s.data[op.Key] = &DataTypeValue.StringValue{Data: op.Value}
		case "SADD":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.SetValue{Data: make(map[string]struct{})}
			}
			if SetValue, ok := s.data[op.Key].(*DataTypeValue.SetValue); ok {
				SetValue.Data[op.Value] = struct{}{}
			}
		case "SPOP":
			if val, exists := s.data[op.Key]; exists {
				if SetValue, ok := val.(*DataTypeValue.SetValue); ok {
					delete(SetValue.Data, op.Value)
				}
			}

		case "LPUSH":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.ListValue{Data: make([]string, 0)}
			}
			if ListValue, ok := s.data[op.Key].(*DataTypeValue.ListValue); ok {
				ListValue.Data = append([]string{op.Value}, ListValue.Data...)
			}
		case "RPUSH":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.ListValue{Data: make([]string, 0)}
			}
			if ListValue, ok := s.data[op.Key].(*DataTypeValue.ListValue); ok {
				ListValue.Data = append(ListValue.Data, op.Value)
			}

		case "ENQUEUE":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.QueueValue{Data: make([]string, 0)}
			}
			if queueValue, ok := s.data[op.Key].(*DataTypeValue.QueueValue); ok {
				queueValue.Data = append(queueValue.Data, op.Value)
			}

		case "DEQUEUE":
			if val, exists := s.data[op.Key]; exists {
				if queueValue, ok := val.(*DataTypeValue.QueueValue); ok {
					if len(queueValue.Data) > 0 {
						queueValue.Data = queueValue.Data[1:]
					}
				}
			}

		case "PUSH":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.StackValue{Data: make([]string, 0)}
			}

			if val, ok := s.data[op.Key].(*DataTypeValue.StackValue); ok {
				val.Data = append(val.Data, op.Value)
			}

		case "POP":
			if val, exists := s.data[op.Key]; exists {
				if val, ok := val.(*DataTypeValue.StackValue); ok && len(val.Data) > 0 {
					val.Data = val.Data[:len(val.Data)-1]
				}
			}

		case "HSET":
			if _, exists := s.data[op.Key]; !exists {
				s.data[op.Key] = &DataTypeValue.HashmapValue{Data: make(map[string]string)}
			}
			if hashVal, ok := s.data[op.Key].(*DataTypeValue.HashmapValue); ok {
				var payload HSetPayload
				if err := json.Unmarshal([]byte(op.Value), &payload); err == nil {
					hashVal.Data[payload.Field] = payload.Value
				} else {
					parts := strings.SplitN(op.Value, ":", 2)
					if len(parts) == 2 {
						hashVal.Data[parts[0]] = parts[1]
					}
				}
			}

		case "DELETE":
			delete(s.data, op.Key)
		}
	}
	logger.Info("AOF loaded successfully")
	return nil
}

// Snapshot triggers AOF snapshot generation
func (s *Store) Snapshot() error {
	if !s.enableAof {
		return fmt.Errorf("AOF is not enabled")
	}
	return s.aof.Snapshot(s)
}
