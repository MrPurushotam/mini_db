package domain

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
