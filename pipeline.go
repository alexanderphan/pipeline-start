package main

import (
	"context"
	"errors"
)

// Message is the envelope that flows through the pipeline.
type Message struct {
	Type    string
	Version string
	Payload map[string]any
	Meta    map[string]string
}

// Source produces messages.
type Source interface {
	Start(ctx context.Context) (<-chan Message, <-chan error)
}

// Stage validates/transforms a message.
type Stage interface {
	Process(ctx context.Context, msg Message) (Message, error)
}

// Destination consumes final messages.
type Destination interface {
	Consume(ctx context.Context, msg Message) error
}

// Pipeline wires Source -> []Stage -> Destination.
type Pipeline struct {
	Source Source
	Stages []Stage
	Sink   Destination
}

// Run executes Source -> Stages -> Destination.
func (p *Pipeline) Run(ctx context.Context) error {
	// Exit immediately if cancellation already happened.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Source and sink are required for a runnable pipeline.
	if p.Source == nil {
		return errors.New("pipeline source is required")
	}
	if p.Sink == nil {
		return errors.New("pipeline sink is required")
	}

	// Start source streaming and read until both channels are closed.
	msgCh, errCh := p.Source.Start(ctx)
	for msgCh != nil || errCh != nil {
		select {
		// Global shutdown path.
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			// Fail fast on source error.
			if err != nil {
				return err
			}
		case msg, ok := <-msgCh:
			if !ok {
				msgCh = nil
				continue
			}

			// Apply stages in sequence; output of one becomes input to the next.
			current := msg
			for _, stage := range p.Stages {
				if err := ctx.Err(); err != nil {
					return err
				}

				next, err := stage.Process(ctx, current)
				if err != nil {
					return err
				}
				current = next
			}

			// Send fully processed message to destination.
			if err := ctx.Err(); err != nil {
				return err
			}

			if err := p.Sink.Consume(ctx, current); err != nil {
				return err
			}
		}
	}

	return nil
}
