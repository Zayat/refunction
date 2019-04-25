package messages

import (
	"context"

	"github.com/segmentio/kafka-go"
)

//go:generate gobin -m -run github.com/maxbrunsfeld/counterfeiter/v6 . Writer

type Writer interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

type writer struct {
	kafkaWriter *kafka.Writer
}

func NewWriter(host string, topic string) Writer {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{host},
		Topic:   topic,
	})
	return &writer{
		kafkaWriter: kafkaWriter,
	}
}

func (p writer) WriteMessages(ctx context.Context, messages ...kafka.Message) error {
	return p.kafkaWriter.WriteMessages(ctx, messages...)
}