package main

import (
	"context"
	"fmt"
)

// StdoutDestination prints each message.
type StdoutDestination struct{}

// Consume prints the message.
func (d *StdoutDestination) Consume(ctx context.Context, msg Message) error {
	_ = ctx

	_, err := fmt.Println(msg)
	return err
}
