package main

import "context"

// SetMetaStage adds one metadata key/value to each message.
type SetMetaStage struct {
	Key   string
	Value string
}

// Process applies the metadata change and returns the updated message.
func (s *SetMetaStage) Process(ctx context.Context, msg Message) (Message, error) {
	_ = ctx
	if msg.Meta == nil {
		msg.Meta = map[string]string{}
	}
	msg.Meta[s.Key] = s.Value
	return msg, nil
}
