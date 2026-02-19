package main

import (
	"context"
	"log"
)

func main() {
	source := &SliceSource{
		Messages: []Message{
			{
				Type:    "user.created",
				Version: "v1",
				Payload: map[string]any{"id": "u_123", "email": "alice@example.com"},
				Meta:    map[string]string{"source": "demo"},
			},
			{
				Type:    "user.updated",
				Version: "v1",
				Payload: map[string]any{"id": "u_123", "email": "alice+new@example.com"},
				Meta:    map[string]string{"source": "demo"},
			},
		},
	}

	validate := &ValidationStage{
		Schemas: map[string]map[string]Schema{
			"user.created": {
				"v1": {RequiredFields: []string{"id", "email"}},
			},
			"user.updated": {
				"v1": {RequiredFields: []string{"id", "email"}},
			},
		},
	}
	transform := &SetMetaStage{Key: "processed_by", Value: "demo-transform"}

	pipeline := &Pipeline{
		Source: source,
		Stages: []Stage{validate, transform},
		Sink:   &StdoutDestination{},
	}

	if err := pipeline.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
