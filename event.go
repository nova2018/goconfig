package goconfig

type ConfigUpdateEvent struct {
	fullKey string
	key     string
	op      int8
}

func (c ConfigUpdateEvent) FullKey() string {
	return c.fullKey
}

func (c ConfigUpdateEvent) Key() string {
	return c.key
}

func (c ConfigUpdateEvent) Op() int8 {
	return c.op
}

const (
	EventOpAdd    = 1
	EventOpUpdate = 2
	EventOpDelete = 3
)
