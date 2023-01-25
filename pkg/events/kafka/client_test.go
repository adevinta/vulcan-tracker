/*
Copyright 2022 Adevinta
*/
package kafka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/adevinta/vulcan-tracker/pkg/events"
	"github.com/adevinta/vulcan-tracker/pkg/events/streamtest"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
)

const (
	bootstrapServers = "127.0.0.1:9092"
	groupPrefix      = "stream_kafka_kafka_test_group_"
	topicPrefix      = "stream_kafka_kafka_test_topic_"
	messagesFile     = "testdata/messages.json"
	timeout          = 5 * time.Minute
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func setupKafka(topic string) (msgs []events.Message, err error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,

		// Set message timeout to 5s, so the kafka client returns an
		// error if the broker is not up.
		"message.timeout.ms": 5000,
	}

	prod, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating producer: %w", err)
	}
	defer prod.Close()

	msgs = streamtest.MustParse(messagesFile)
	for _, msg := range msgs {
		if err := produceMessage(prod, topic, msg); err != nil {
			return nil, fmt.Errorf("error producing message: %w", err)
		}
	}

	return msgs, nil
}

func produceMessage(prod *kafka.Producer, topic string, msg events.Message) error {
	events := make(chan kafka.Event)
	defer close(events)

	kmsg := &kafka.Message{
		Key:            msg.Key,
		Value:          msg.Value,
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
	}

	for _, e := range msg.Headers {
		hdr := kafka.Header{
			Key:   string(e.Key),
			Value: e.Value,
		}
		kmsg.Headers = append(kmsg.Headers, hdr)
	}

	if err := prod.Produce(kmsg, events); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-events
	kmsg, ok := e.(*kafka.Message)
	if !ok {
		return errors.New("event type is not *events.Message")
	}
	if kmsg.TopicPartition.Error != nil {
		return fmt.Errorf("could not deliver message: %w", kmsg.TopicPartition.Error)
	}

	return nil
}

func TestProcessorProcess(t *testing.T) {
	topic := topicPrefix + strconv.FormatInt(rand.Int63(), 16)

	want, err := setupKafka(topic)
	if err != nil {
		t.Fatalf("error setting up kafka: %v", err)
	}

	cfg := map[string]any{
		"bootstrap.servers":       bootstrapServers,
		"group.id":                groupPrefix + strconv.FormatInt(rand.Int63(), 16),
		"auto.commit.interval.ms": 100,
		"auto.offset.reset":       "earliest",
	}

	proc, err := NewProcessor(cfg)
	if err != nil {
		t.Fatalf("error creating kafka processor: %v", err)
	}
	defer proc.Close()

	var (
		ctr int
		got []events.Message
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = proc.Process(ctx, topic, func(msg events.Message) error {
		got = append(got, msg)

		ctr++
		if ctr >= len(want) {
			cancel()
		}

		return nil
	})
	if err != nil {
		t.Fatalf("error processing messages: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("messages mismatch (-want +got):\n%v", diff)
	}
}

func TestProcessorProcessAtLeastOnce(t *testing.T) {
	// Number of messages to process before error.
	const n = 2

	topic := topicPrefix + strconv.FormatInt(rand.Int63(), 16)

	want, err := setupKafka(topic)
	if err != nil {
		t.Fatalf("error setting up kafka: %v", err)
	}

	if n > len(want) {
		t.Fatal("n > testdata length")
	}

	cfg := map[string]any{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupPrefix + strconv.FormatInt(rand.Int63(), 16),
		"auto.offset.reset": "earliest",
	}

	proc, err := NewProcessor(cfg)
	if err != nil {
		t.Fatalf("error creating kafka processor: %v", err)
	}
	defer proc.Close()

	var (
		ctr int
		got []events.Message
	)

	// Fail after processing n messages.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = proc.Process(ctx, topic, func(msg events.Message) error {
		if ctr >= n {
			return errors.New("error")
		}

		got = append(got, msg)
		ctr++

		return nil
	})
	if err == nil {
		t.Fatalf("Process should have returned error: %v", err)
	}

	// Wait for 1s to ensure that the offsets are commited.
	time.Sleep(1 * time.Second)

	// Resume stream processing.
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = proc.Process(ctx, topic, func(msg events.Message) error {
		got = append(got, msg)

		ctr++
		if ctr >= len(want) {
			cancel()
		}

		return nil
	})
	if err != nil {
		t.Fatalf("error processing messages: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("messages mismatch (-want +got):\n%v", diff)
	}
}
