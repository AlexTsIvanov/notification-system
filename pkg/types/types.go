package types

type EventContext struct{
	EventId string
	Payload   []byte
	RetryCount int
}