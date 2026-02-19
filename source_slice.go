package main

import "context"

// SliceSource emits a fixed in-memory slice of messages.
type SliceSource struct {
	Messages []Message
}

// Start streams all messages unless the context is cancelled.
func (s *SliceSource) Start(ctx context.Context) (<-chan Message, <-chan error) {
	msgCh := make(chan Message)
	errCh := make(chan error)

	go func() {
		defer close(msgCh)
		defer close(errCh)

		for _, msg := range s.Messages {
			if ctx.Err() != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case msgCh <- msg:
			}
		}
	}()

	return msgCh, errCh
}
