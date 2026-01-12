package value

import (
	"encoding/json"

	"github.com/mrpurushotam/mini_db/internal/domain"
)

//Where value is list type
type ListValue struct {
	Data []string
}

func (l *ListValue) Type() domain.DataType {
	return domain.List
}

func (l *ListValue) Serialize() []byte {
	data, _ := json.Marshal(l.Data)
	return data
}
func (l *ListValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &l.Data)
}
