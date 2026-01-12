package value

import (
	"encoding/json"

	"github.com/mrpurushotam/mini_db/internal/domain"
)

// Where value is queue type
type QueueValue struct {
	Data []string
}

func (q *QueueValue) Type() domain.DataType {
	return domain.Queue
}

func (q *QueueValue) Serialize() []byte {
	data, _ := json.Marshal(q.Data)
	return data
}
func (q *QueueValue) Deserialize(data []byte) error {
	return json.Unmarshal(data, &q.Data)
}
