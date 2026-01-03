package store

import "encoding/json"

type DataType string

const (
	String  DataType = "string"
	Set     DataType = "set"
	List    DataType = "list"
	Queue   DataType = "queue"
	Stack   DataType = "stack"
	Hashmap DataType = "hashmap"
)

type Value interface {
	Type() DataType
	Serialize() []byte
	Deserialize([]byte) error
}

//Where value is string type
type StringValue struct {
	Data string
}

func (s *StringValue) Type() DataType {
	return String
}

func (s *StringValue) Serialize() []byte {
	return []byte(s.Data)
}

func (s *StringValue) Deserialize(data []byte) error {
	s.Data = string(data)
	return nil
}

//Where value is Set type
type SetValue struct {
	Data map[string]struct{}
}

func (s *SetValue) Type() DataType {
	return Set
}

func (s *SetValue) Serialize() []byte {
	members := make([]string, 0, len(s.Data))
	for member := range s.Data {
		members = append(members, member)
	}
	data, _ := json.Marshal(members)
	return data
}

func (s *SetValue) Deserialize(data []byte) error {
	var members []string
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}
	s.Data = make(map[string]struct{})
	for _, member := range members {
		s.Data[member] = struct{}{}
	}
	return nil
}

//Where value is Stack type
type StackValue struct {
	Data []string
}

func (s *StackValue) Type() DataType {
	return Stack
}

func (s *StackValue) Serialize() []byte {
	data, _ := json.Marshal(s.Data)
	return data
}

func (s *StackValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &s.Data)

}

//Where value is list type
type ListValue struct {
	Data []string
}

func (l *ListValue) Type() DataType {
	return List
}

func (l *ListValue) Serialize() []byte {
	data, _ := json.Marshal(l.Data)
	return data
}
func (l *ListValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &l.Data)
}

//Where value is queue type
type QueueValue struct {
	Data []string
}

func (q *QueueValue) Type() DataType {
	return Queue
}

func (q *QueueValue) Serialize() []byte {
	data, _ := json.Marshal(q.Data)
	return data
}
func (q *QueueValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &q.Data)
}

//Where value is hashmap type
type HashmapValue struct {
	Data map[string]string
}

func (h *HashmapValue) Type() DataType {
	return Hashmap
}

func (h *HashmapValue) Serialize() []byte {
	data, _ := json.Marshal(h.Data)
	return data
}
func (h *HashmapValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &h.Data)
}
