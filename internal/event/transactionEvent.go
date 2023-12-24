package event

type TransactionEvent struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}

type EventType byte

const (
	_                  = iota
	EventPut EventType = iota
	EventDelete
)
