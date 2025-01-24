package domain

type DeserializationFmt uint

const (
	JSON DeserializationFmt = iota
	Protobuf
)

type Config struct {
	Topic         string
	ConsumerGroup string
	DeserFmt      DeserializationFmt
}
