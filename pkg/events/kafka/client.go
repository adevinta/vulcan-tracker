/*
Copyright 2022 Adevinta
*/
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/adevinta/vulcan-tracker/pkg/events"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// A Processor allows to process messages from a kafka topic.
type Processor struct {
	c *kafka.Consumer
}

// NewProcessor returns a [Processor] with the provided kafka
// configuration properties.
func NewProcessor(config map[string]any) (Processor, error) {
	kconfig := make(kafka.ConfigMap)
	for k, v := range config {
		if err := kconfig.SetKey(k, v); err != nil {
			return Processor{}, fmt.Errorf("could not set config key: %w", err)
		}
	}
	kconfig["enable.auto.commit"] = false
	kconfig["enable.auto.offset.store"] = false

	c, err := kafka.NewConsumer(&kconfig)
	if err != nil {
		return Processor{}, fmt.Errorf("failed to create a consumer: %w", err)
	}

	return Processor{c}, nil
}

// Process processes the messages received in the topic called entity by
// calling h. This method blocks the calling goroutine until the specified
// context is cancelled or an error occurs. It replaces the current kafka
// subscription, so it should not be called concurrently.
func (proc Processor) Process(ctx context.Context, entity string, h events.MsgHandler) error {
	if err := proc.c.Subscribe(entity, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topic %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		kmsg, err := proc.c.ReadMessage(100 * time.Millisecond)
		if err != nil {
			kerr, ok := err.(kafka.Error)
			if ok && kerr.Code() == kafka.ErrTimedOut {
				continue
			}
			return fmt.Errorf("error reading message: %w", kerr)
		}

		msg := events.Message{
			Key:   kmsg.Key,
			Value: kmsg.Value,
		}

		for _, hdr := range kmsg.Headers {
			entry := events.MetadataEntry{
				Key:   []byte(hdr.Key),
				Value: hdr.Value,
			}
			msg.Headers = append(msg.Headers, entry)
		}

		if err := h(msg); err != nil {
			return fmt.Errorf("error processing message: %w", err)
		}

		if err := proc.Commit(*kmsg); err != nil {
			return fmt.Errorf("error commiting message: %w", err)
		}
	}
}

// Close closes the underlaying kafka consumer.
func (proc Processor) Close() error {
	return proc.c.Close()
}

func (proc Processor) Commit(kmsg kafka.Message) error {
	_, err := proc.c.CommitOffsets([]kafka.TopicPartition{
		{
			Topic:     kmsg.TopicPartition.Topic,
			Partition: kmsg.TopicPartition.Partition,
			Offset:    kmsg.TopicPartition.Offset + 1,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
