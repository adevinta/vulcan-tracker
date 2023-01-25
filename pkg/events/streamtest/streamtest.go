/*
Copyright 2022 Adevinta
*/
// Package streamtest provides utilities for stream testing.
package streamtest

import (
	"context"
	"encoding/json"
	"os"

	"github.com/adevinta/vulcan-tracker/pkg/events"
)

// MustParse parses a json file with messages and returns them. It panics if
// the file cannot be parsed.
func MustParse(filename string) []events.Message {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var testdata []struct {
		Key     *string                     `json:"key,omitempty"`
		Value   *events.FindingNotification `json:"value,omitempty"`
		Headers []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"headers,omitempty"`
	}

	if err := json.NewDecoder(f).Decode(&testdata); err != nil {
		panic(err)
	}

	var msgs []events.Message
	for _, td := range testdata {
		var msg events.Message
		if td.Key != nil {
			msg.Key = []byte(*td.Key)
		}
		if td.Value != nil {
			payload, err := json.Marshal(td.Value)
			if err != nil {
				panic(err)
			}
			msg.Value = payload
		}
		for _, e := range td.Headers {
			if e.Key == "" {
				panic("empty metadata key")
			}
			if e.Value == "" {
				panic("empty metadata value")
			}
			entry := events.MetadataEntry{
				Key:   []byte(e.Key),
				Value: []byte(e.Value),
			}
			msg.Headers = append(msg.Headers, entry)
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

// MockProcessor mocks a stream processor with a predefined set of messages. It
// implements the interface [stream.Processor].
type MockProcessor struct {
	msgs []events.Message
}

// NewMockProcessor returns a [MockProcessor]. It initializes its internal list
// of messages with msgs.
func NewMockProcessor(msgs []events.Message) *MockProcessor {
	return &MockProcessor{msgs}
}

// Process processes the messages passed to [NewMockProcessor].
func (mp *MockProcessor) Process(ctx context.Context, entity string, h events.MsgHandler) error {
	for _, msg := range mp.msgs {
		if err := h(msg); err != nil {
			return err
		}
	}
	return nil
}
