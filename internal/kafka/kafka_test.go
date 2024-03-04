package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/ig0rmin/ich/internal/config"
	"github.com/stretchr/testify/require"
)

type MockSubscirber struct {
	Received [][]byte
}

func (s *MockSubscirber) Receive(data []byte) error {
	s.Received = append(s.Received, data)
	return nil
}

func TestKafka(t *testing.T) {
	var cfg Config
	require.NoError(t, config.Load(&cfg))

	messages, err := NewKafka(cfg, "topic-messages")
	require.NoError(t, err)
	defer messages.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go messages.Run(ctx)

	var subscriber MockSubscirber
	messages.Subscribe(&subscriber)

	var data = []byte("hello from unit tests")
	messages.Publish(data)

	// Give Kafka time to receive the message
	time.Sleep(500 * time.Millisecond)

	cancel()

	messages.Wait()

	require.Equal(t, 1, len(subscriber.Received))
	require.Equal(t, data, subscriber.Received[0])
}
