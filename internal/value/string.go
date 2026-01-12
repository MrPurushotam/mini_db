package value

import (
	"github.com/mrpurushotam/mini_db/internal/domain"
)

// Where value is string type
type StringValue struct {
	Data string
}

func (s *StringValue) Type() domain.DataType {
	return domain.String
}

func (s *StringValue) Serialize() []byte {
	return []byte(s.Data)
}

func (s *StringValue) Deserialize(data []byte) error {
	s.Data = string(data)
	return nil
}
