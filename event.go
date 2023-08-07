package goconfig

type ConfigUpdateEvent struct {
	key    string
	subKey string
	op     int8
}

func (c ConfigUpdateEvent) Key() string {
	return c.key
}

func (c ConfigUpdateEvent) SubKey() string {
	return c.subKey
}

func (c ConfigUpdateEvent) Op() int8 {
	return c.op
}

const (
	EventOpAdd    = 1
	EventOpUpdate = 2
	EventOpDelete = 3
)
