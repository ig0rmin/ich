package users

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/kafka"
)

type UserEventsListener interface {
	ReceiveUserJoined(msg *api.UserJoinedMsg)
	ReceiveUserLeft(msg *api.UserLeftMsg)
}

type UserManager struct {
	users *kafka.Kafka

	// Users on this server
	localUsers      map[string]struct{}
	localUsersMutex sync.Mutex

	// All users (from all servers)
	usersOnline      map[string]struct{}
	usersOnlineMutex sync.Mutex

	listeners      map[UserEventsListener]struct{}
	listenersMutex sync.Mutex
}

func NewUserManager(users *kafka.Kafka) (*UserManager, error) {
	u := &UserManager{
		users:       users,
		usersOnline: make(map[string]struct{}),
		localUsers:  make(map[string]struct{}),
		listeners:   make(map[UserEventsListener]struct{}),
	}
	return u, nil
}

func (u *UserManager) Init() {
	u.users.Subscribe(u)
	// Make all servers advertise their users to build the list of users online
	u.notifyServerJoined()
}

func (u *UserManager) Close() {
	u.users.Unsubscribe(u)
}

func (u *UserManager) Subscribe(l UserEventsListener) {
	u.listenersMutex.Lock()
	u.listeners[l] = struct{}{}
	u.listenersMutex.Unlock()
}

func (u *UserManager) Unsubscribe(l UserEventsListener) {
	u.listenersMutex.Lock()
	delete(u.listeners, l)
	u.listenersMutex.Unlock()
}

func (u *UserManager) Receive(data []byte) error {
	var msg api.Msg
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case api.TypeServerJoined:
		u.onServerJoined()
	case api.TypeUserJoined:
		u.onUserJoined(msg.Msg)
	case api.TypeUserLeft:
		u.onUserLeft(msg.Msg)
	case api.TypeUsersOnline:
		u.onUsersOnline(msg.Msg)
	default:
		log.Error("Unsupported message type")
	}
	return nil
}

// When a new server joins, each server online advertises its users
// so the newcomer can build the list of users online
func (u *UserManager) onServerJoined() error {
	users := &api.UsersOnline{
		List: make([]string, 0, len(u.localUsers)),
	}
	u.localUsersMutex.Lock()
	for name := range u.localUsers {
		users.List = append(users.List, name)
	}
	u.localUsersMutex.Unlock()

	msg := &api.Msg{
		Type:   api.TypeUsersOnline,
		SentAt: time.Now(),
		Msg:    &users,
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) onUserJoined(msg any) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var user api.UserJoinedMsg
	if err := json.Unmarshal(rawMsg, &user); err != nil {
		return err
	}
	u.usersOnlineMutex.Lock()
	u.usersOnline[user.UserName] = struct{}{}
	u.usersOnlineMutex.Unlock()

	u.listenersMutex.Lock()
	for l := range u.listeners {
		l.ReceiveUserJoined(&user)
	}
	u.listenersMutex.Unlock()
	return nil
}

func (u *UserManager) onUserLeft(msg any) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var user api.UserLeftMsg
	if err := json.Unmarshal(rawMsg, &user); err != nil {
		return err
	}
	u.usersOnlineMutex.Lock()
	delete(u.usersOnline, user.UserName)
	u.usersOnlineMutex.Unlock()

	u.listenersMutex.Lock()
	for l := range u.listeners {
		l.ReceiveUserLeft(&user)
	}
	u.listenersMutex.Unlock()
	return nil
}

func (u *UserManager) onUsersOnline(msg any) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var users api.UsersOnline
	if err := json.Unmarshal(rawMsg, &users); err != nil {
		return err
	}
	u.usersOnlineMutex.Lock()
	for _, name := range users.List {
		u.usersOnline[name] = struct{}{}
	}
	u.usersOnlineMutex.Unlock()
	return nil
}

func (u *UserManager) GetUsersOnline() []string {
	res := make([]string, 0, len(u.usersOnline))
	u.usersOnlineMutex.Lock()
	for name := range u.usersOnline {
		res = append(res, name)
	}
	u.usersOnlineMutex.Unlock()
	return res
}

func (u *UserManager) notifyServerJoined() error {
	msg := &api.Msg{
		Type:   "server_joined",
		SentAt: time.Now(),
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) NotifyUserJoined(userName string) error {
	u.localUsersMutex.Lock()
	u.localUsers[userName] = struct{}{}
	u.localUsersMutex.Unlock()

	msg := &api.Msg{
		Type:   "user_joined",
		SentAt: time.Now(),
		Msg: &api.UserJoinedMsg{
			UserName: userName,
		},
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) NotifyUserLeft(userName string) error {
	u.localUsersMutex.Lock()
	u.localUsers[userName] = struct{}{}
	u.localUsersMutex.Unlock()

	msg := &api.Msg{
		Type:   "user_left",
		SentAt: time.Now(),
		Msg: &api.UserLeftMsg{
			UserName: userName,
		},
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) pulbishMsg(msg *api.Msg) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	u.users.Publish(rawMsg)
	return nil
}
