package value

import (
	"encoding/json"

	"github.com/mrpurushotam/mini_db/internal/domain"
)

// Where value is Stack type
type StackValue struct {
	Data []string
}

func (s *StackValue) Type() domain.DataType {
	return domain.Stack
}

func (s *StackValue) Serialize() []byte {
	data, _ := json.Marshal(s.Data)
	return data
}

func (s *StackValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &s.Data)

}
