package messages

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/kafka"
)

type MessageListener interface {
	ReceiveChatMessage(time.Time, *api.ChatMessage)
}

type Messages struct {
	messages *kafka.Kafka

	listeners      map[MessageListener]struct{}
	listenersMutex sync.Mutex
}

func NewMessages(messages *kafka.Kafka) (*Messages, error) {
	return &Messages{
		messages:  messages,
		listeners: make(map[MessageListener]struct{}),
	}, nil
}

func (m *Messages) Subscribe(l MessageListener) {
	m.listenersMutex.Lock()
	m.listeners[l] = struct{}{}
	m.listenersMutex.Unlock()
}

func (m *Messages) Unsubscribe(l MessageListener) {
	m.listenersMutex.Lock()
	delete(m.listeners, l)
	m.listenersMutex.Unlock()
}

func (m *Messages) Init() {
	m.messages.Subscribe(m)
}

func (m *Messages) Close() {
	m.messages.Unsubscribe(m)
}

func (m *Messages) PostChatMessage(chatMsg *api.ChatMessage) error {
	msg := &api.Msg{
		Type:   api.TypeChatMessage,
		SentAt: time.Now(),
		Msg:    chatMsg,
	}
	msgRaw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	m.messages.Publish(msgRaw)
	return nil
}

func (u *Messages) Receive(data []byte) error {
	var msg api.Msg
	chatMsg := &api.ChatMessage{}
	msg.Msg = chatMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil
	}
	if msg.Type != api.TypeChatMessage {
		return fmt.Errorf("unsupported message type: %v", msg.Type)
	}
	u.notifyListeners(msg.SentAt, chatMsg)
	return nil
}

func (m *Messages) notifyListeners(sentAt time.Time, msg *api.ChatMessage) {
	m.listenersMutex.Lock()
	for l := range m.listeners {
		l.ReceiveChatMessage(sentAt, msg)
	}
	m.listenersMutex.Unlock()
}
