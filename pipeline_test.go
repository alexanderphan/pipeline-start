package main

import (
	"context"
	"errors"
	"testing"
)

var (
	errStage  = errors.New("stage failed")
)

type failingStage struct{}

func (s *failingStage) Process(ctx context.Context, msg Message) (Message, error) {
	_ = ctx
	_ = msg
	return Message{}, errStage
}

// TestSliceSourceEmitsAllMessages verifies the source emits all configured messages.
func TestSliceSourceEmitsAllMessages(t *testing.T) {
	src := &SliceSource{
		Messages: []Message{
			{Type: "a", Version: "v1"},
			{Type: "b", Version: "v1"},
		},
	}

	msgCh, errCh := src.Start(context.Background())

	var got []Message
	for msg := range msgCh {
		got = append(got, msg)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	if got[0].Type != "a" || got[1].Type != "b" {
		t.Fatalf("unexpected message order: %+v", got)
	}
	if _, ok := <-errCh; ok {
		t.Fatal("expected closed error channel")
	}
}

// TestValidationStageUsesTypeAndVersionSchema verifies schema lookup by type+version.
func TestValidationStageUsesTypeAndVersionSchema(t *testing.T) {
	stage := &ValidationStage{
		Schemas: map[string]map[string]Schema{
			"user.created": {
				"v1": {RequiredFields: []string{"id", "email"}},
			},
		},
	}

	_, err := stage.Process(context.Background(), Message{
		Type:    "user.created",
		Version: "v1",
		Payload: map[string]any{"id": "u_1", "email": "a@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error for valid message: %v", err)
	}

	_, err = stage.Process(context.Background(), Message{
		Type:    "user.created",
		Version: "v2",
		Payload: map[string]any{"id": "u_1", "email": "a@example.com"},
	})
	if !errors.Is(err, ErrSchemaNotFound) {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}

	_, err = stage.Process(context.Background(), Message{
		Type:    "user.created",
		Version: "v1",
		Payload: map[string]any{"id": "u_1"},
	})
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("expected ErrMissingField, got %v", err)
	}
}

// TestSetMetaStageAddsMetadata verifies the transform stage mutates metadata.
func TestSetMetaStageAddsMetadata(t *testing.T) {
	stage := &SetMetaStage{Key: "processed_by", Value: "set-meta"}

	msg, err := stage.Process(context.Background(), Message{Type: "x", Version: "v1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Meta["processed_by"] != "set-meta" {
		t.Fatalf("expected metadata key to be set, got %+v", msg.Meta)
	}
}

// TestMemoryDestinationStoresMessages verifies destination consumption in isolation.
func TestMemoryDestinationStoresMessages(t *testing.T) {
	d := &MemoryDestination{}

	if err := d.Consume(context.Background(), Message{Type: "x", Version: "v1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(d.Messages))
	}
}

// TestPipelineRunWithValidationAndTransform verifies end-to-end happy path.
func TestPipelineRunWithValidationAndTransform(t *testing.T) {
	source := &SliceSource{
		Messages: []Message{
			{Type: "user.created", Version: "v1", Payload: map[string]any{"id": "u_1", "email": "a@example.com"}},
		},
	}
	validate := &ValidationStage{
		Schemas: map[string]map[string]Schema{
			"user.created": {
				"v1": {RequiredFields: []string{"id", "email"}},
			},
		},
	}
	transform := &SetMetaStage{Key: "processed_by", Value: "set-meta"}
	dest := &MemoryDestination{}

	p := &Pipeline{Source: source, Stages: []Stage{validate, transform}, Sink: dest}

	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dest.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(dest.Messages))
	}
	if dest.Messages[0].Meta["processed_by"] != "set-meta" {
		t.Fatalf("expected transformed metadata, got %+v", dest.Messages[0].Meta)
	}
}

// TestPipelineRunReturnsErrorWhenSourceMissing verifies required source check.
func TestPipelineRunReturnsErrorWhenSourceMissing(t *testing.T) {
	p := &Pipeline{Sink: &MemoryDestination{}}
	err := p.Run(context.Background())
	if err == nil || err.Error() != "pipeline source is required" {
		t.Fatalf("expected missing source error, got %v", err)
	}
}

// TestPipelineRunReturnsErrorWhenSinkMissing verifies required sink check.
func TestPipelineRunReturnsErrorWhenSinkMissing(t *testing.T) {
	p := &Pipeline{Source: &SliceSource{}}
	err := p.Run(context.Background())
	if err == nil || err.Error() != "pipeline sink is required" {
		t.Fatalf("expected missing sink error, got %v", err)
	}
}

// TestPipelineRunReturnsStageError verifies fail-fast behavior on stage errors.
func TestPipelineRunReturnsStageError(t *testing.T) {
	p := &Pipeline{
		Source: &SliceSource{Messages: []Message{{Type: "x", Version: "v1"}}},
		Stages: []Stage{&failingStage{}},
		Sink:   &MemoryDestination{},
	}
	if err := p.Run(context.Background()); !errors.Is(err, errStage) {
		t.Fatalf("expected stage error, got %v", err)
	}
}

// TestPipelineRunStopsWhenContextCancelled verifies graceful termination on cancellation.
func TestPipelineRunStopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &Pipeline{Source: &SliceSource{Messages: []Message{{Type: "x", Version: "v1"}}}, Sink: &MemoryDestination{}}
	if err := p.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
