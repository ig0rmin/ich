package messages

import (
	"context"
	"testing"
	"time"

	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/config"
	"github.com/ig0rmin/ich/internal/kafka"
	"github.com/stretchr/testify/require"
)

type MockServer struct {
	topic    *kafka.Kafka
	messages *Messages
}

func NewMockServer(t *testing.T, ctx context.Context) *MockServer {
	var cfg kafka.Config
	require.NoError(t, config.Load(&cfg))

	topic, err := kafka.NewKafka(cfg, "topic-messages")
	require.NoError(t, err)

	go topic.Run(ctx)

	m, err := NewMessages(topic)
	require.NoError(t, err)

	m.Init()

	return &MockServer{
		topic:    topic,
		messages: m,
	}
}

func (m *MockServer) Close() {
	m.topic.Wait()
	m.topic.Close()
}

type MockMessageListener struct {
	Received []api.ChatMessage
}

func (m *MockMessageListener) ReceiveChatMessage(sentAt time.Time, chatMsg *api.ChatMessage) {
	m.Received = append(m.Received, *chatMsg)
}

func TestMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := NewMockServer(t, ctx)
	defer server.Close()
	defer cancel()

	server.messages.Init()

	var listener MockMessageListener
	server.messages.Subscribe(&listener)

	msg := &api.ChatMessage{
		From: "Patrick",
		Text: "Hello!",
	}

	server.messages.PostChatMessage(msg)

	// Let Kafka time to process messages
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, 1, len(listener.Received))
	require.Contains(t, listener.Received, *msg)

	server.messages.Close()
}
