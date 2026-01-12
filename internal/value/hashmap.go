package value

import (
	"encoding/json"

	"github.com/mrpurushotam/mini_db/internal/domain"
)

// Where value is hashmap type
type HashmapValue struct {
	Data map[string]string
}

func (h *HashmapValue) Type() domain.DataType {
	return domain.Hashmap
}

func (h *HashmapValue) Serialize() []byte {
	data, _ := json.Marshal(h.Data)
	return data
}
func (h *HashmapValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &h.Data)
}
