package value

import (
	"encoding/json"

	"github.com/mrpurushotam/mini_db/internal/domain"
)

// Where value is Set type
type SetValue struct {
	Data map[string]struct{}
}

func (s *SetValue) Type() domain.DataType {
	return domain.Set
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
