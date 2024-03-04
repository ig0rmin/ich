package users

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/config"
	"github.com/ig0rmin/ich/internal/kafka"
	"github.com/stretchr/testify/require"
)

type MockServer struct {
	users       *kafka.Kafka
	userManager *UserManager
}

func (m *MockServer) Close() {
	m.users.Wait()
	m.users.Close()
}

func NewMockServer(t *testing.T, ctx context.Context) *MockServer {
	os.Chdir("../..")
	var cfg kafka.Config
	require.NoError(t, config.Load(&cfg))

	users, err := kafka.NewKafka(cfg, "topic-users")
	require.NoError(t, err)

	go users.Run(ctx)

	um, err := NewUserManager(users)
	require.NoError(t, err)

	um.Init()

	return &MockServer{
		users:       users,
		userManager: um,
	}
}

type MockUsersListener struct {
	Joined []string
	Left   []string
}

func (l *MockUsersListener) ReceiveUserJoined(msg *api.UserJoinedMsg) {
	l.Joined = append(l.Joined, msg.UserName)
}

func (l *MockUsersListener) ReceiveUserLeft(msg *api.UserLeftMsg) {
	l.Left = append(l.Left, msg.UserName)
}

func TestUserManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	server1 := NewMockServer(t, ctx)
	defer server1.Close()
	defer cancel()

	var userListener MockUsersListener
	server1.userManager.Subscribe(&userListener)

	server1.userManager.NotifyUserJoined("Spongebob")
	server1.userManager.NotifyUserJoined("Patrick")

	// Let Kafka time to process messages
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, 2, len(server1.userManager.GetUsersOnline()))
	require.Contains(t, server1.userManager.GetUsersOnline(), "Spongebob")
	require.Contains(t, server1.userManager.GetUsersOnline(), "Patrick")

	server2 := NewMockServer(t, ctx)
	defer server2.Close()
	defer cancel()

	// Let Kafka time to process messages
	time.Sleep(500 * time.Millisecond)

	// server2 know about users from server1
	require.Equal(t, 2, len(server2.userManager.GetUsersOnline()))
	require.Contains(t, server2.userManager.GetUsersOnline(), "Spongebob")
	require.Contains(t, server2.userManager.GetUsersOnline(), "Patrick")

	// Spongebob leaves
	server1.userManager.NotifyUserLeft("Spongebob")

	// Let Kafka time to process messages
	time.Sleep(500 * time.Millisecond)

	// server1 has the correct list of users
	require.Equal(t, 1, len(server2.userManager.GetUsersOnline()))
	require.Contains(t, server2.userManager.GetUsersOnline(), "Patrick")

	// server2 has the correct list of users
	require.Equal(t, 1, len(server2.userManager.GetUsersOnline()))
	require.Contains(t, server2.userManager.GetUsersOnline(), "Patrick")

	// Check that listener received all the events
	require.Equal(t, 2, len(userListener.Joined))
	require.Contains(t, userListener.Joined, "Patrick")
	require.Contains(t, userListener.Joined, "Spongebob")
	require.Equal(t, 1, len(userListener.Left))
	require.Contains(t, userListener.Left, "Spongebob")

	// Make the UserManager unsubscribe from events
	server1.userManager.Close()
	server2.userManager.Close()
}
