package main

import (
	"context"
	"errors"
	"fmt"
)

var ErrSchemaNotFound = errors.New("schema not found")
var ErrMissingField = errors.New("required payload field missing")

// Schema is a minimal schema definition for a message.
type Schema struct {
	RequiredFields []string
}

// ValidationStage validates messages by type+version schema.
type ValidationStage struct {
	Schemas map[string]map[string]Schema
}

// Process validates a message using the schema chosen by type and version.
func (s *ValidationStage) Process(ctx context.Context, msg Message) (Message, error) {
	_ = ctx

	versions, ok := s.Schemas[msg.Type]
	if !ok {
		return Message{}, fmt.Errorf("%w: type=%q version=%q", ErrSchemaNotFound, msg.Type, msg.Version)
	}
	schema, ok := versions[msg.Version]
	if !ok {
		return Message{}, fmt.Errorf("%w: type=%q version=%q", ErrSchemaNotFound, msg.Type, msg.Version)
	}

	for _, field := range schema.RequiredFields {
		if msg.Payload == nil {
			return Message{}, fmt.Errorf("%w: %s", ErrMissingField, field)
		}
		if _, exists := msg.Payload[field]; !exists {
			return Message{}, fmt.Errorf("%w: %s", ErrMissingField, field)
		}
	}

	return msg, nil
}
