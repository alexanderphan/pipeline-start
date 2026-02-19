package main

import (
	"context"
)

// MemoryDestination stores consumed messages in memory.
type MemoryDestination struct {
	Messages []Message
}

// Consume appends the message to memory.
func (d *MemoryDestination) Consume(ctx context.Context, msg Message) error {
	_ = ctx

	d.Messages = append(d.Messages, msg)
	return nil
}
