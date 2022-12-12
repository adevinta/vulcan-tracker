/*
Copyright 2022 Adevinta
*/
package vulcan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/adevinta/vulcan-tracker/pkg/events"
)

const (
	// MajorVersion is the major version of the vulnerability db asynchronous API
	// supported by [Client].
	MajorVersion = 0

	// FindingEntityName is the name of the entity linked to findings.
	FindingEntityName = "findings-v0"
)

// Client is a vulnerability db async API client.
type Client struct {
	proc events.Processor
}

var ErrUnsupportedVersion = errors.New("unsupported version")

// NewClient returns a client for the vulnerability db async API using the provided
// stream processor.
func NewClient(proc events.Processor) Client {
	return Client{proc}
}

// FindingHandler processes a finding. isNil is true when the value of the stream
// message is nil.
type FindingHandler func(payload events.FindingNotification, isNil bool) error

// ProcessFindings receives findigns from the underlying stream and processes them
// using the provided handler. This method blocks the calling goroutine until
// the specified context is cancelled.
func (c Client) ProcessFindings(ctx context.Context, h FindingHandler) error {
	return c.proc.Process(ctx, FindingEntityName, func(msg events.Message) error {
		version, err := parseMetadata(msg)
		if err != nil {
			return fmt.Errorf("invalid metadata: %w", err)
		}

		if !supportedVersion(version) {
			return ErrUnsupportedVersion
		}

		id := string(msg.Key)

		var (
			payload events.FindingNotification
			isNil   bool
		)

		if msg.Value != nil {
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				return fmt.Errorf("could not unmarshal finding with ID %q: %w", id, err)
			}
		}

		return h(payload, isNil)
	})
}

// parseMetadata parses and validates message metadata.
func parseMetadata(msg events.Message) (version string, err error) {
	for _, e := range msg.Headers {
		key := string(e.Key)
		value := string(e.Value)

		switch key {
		case "version":
			version = value
		}
	}

	if version == "" {
		return "", errors.New("missing metadata entry")
	}

	return version, nil
}

// supportedVersion takes a semantic version string and returns true if it is
// compatible with [Client].
func supportedVersion(v string) bool {
	if v == "" {
		return false
	}

	if v[0] == 'v' {
		v = v[1:]
	}

	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return major == MajorVersion
}
