package users

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/ig0rmin/ich/internal/kafka"
)

type UserManager struct {
	users *kafka.Kafka

	// Users on this server
	localUsers      map[string]struct{}
	localUsersMutex sync.Mutex

	// All users (from all servers)
	usersOnline      map[string]struct{}
	usersOnlineMutex sync.Mutex
}

type Msg struct {
	Type   string    `json:"type" binding:"required"`
	SentAt time.Time `json:"sent_at"`
	Msg    any       `json:"msg,omitempty" binding:"required,omitempty"`
}

type UserJoinedMsg struct {
	UserName string `json:"username"`
}

type UserLeftMsg struct {
	UserName string `json:"username"`
}

type UsersOnline struct {
	List []string `json:"list"`
}

const (
	TypeServerJoined = "server_joined"
	TypeUserJoined   = "user_joined"
	TypeUserLeft     = "user_left"
	TypeUsersOnline  = "users_online"
)

func NewUserManager(users *kafka.Kafka) (*UserManager, error) {
	u := &UserManager{
		users:       users,
		usersOnline: make(map[string]struct{}),
		localUsers:  make(map[string]struct{}),
	}
	u.users.Subscribe(u)
	// Make all servers advertise their users to build the list of users online
	u.notifyServerJoined()
	return u, nil
}

func (u *UserManager) Close() {
	u.users.Unsubscribe(u)
}

func (u *UserManager) Receive(data []byte) error {
	var msg Msg
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case TypeServerJoined:
		u.onServerJoined()
	case TypeUserJoined:
		u.onUserJoined(msg.Msg)
	case TypeUserLeft:
		u.onUserLeft(msg.Msg)
	case TypeUsersOnline:
		u.onUsersOnline(msg.Msg)
	}
	return nil
}

// When a new server joins, each server online advertises its users
// so the newcomer can build the list of users online
func (u *UserManager) onServerJoined() error {
	users := &UsersOnline{
		List: make([]string, 0, len(u.localUsers)),
	}
	u.localUsersMutex.Lock()
	for name := range u.localUsers {
		users.List = append(users.List, name)
	}
	u.localUsersMutex.Unlock()

	msg := &Msg{
		Type:   TypeUsersOnline,
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
	var user UserJoinedMsg
	if err := json.Unmarshal(rawMsg, &user); err != nil {
		return err
	}
	u.usersOnlineMutex.Lock()
	u.usersOnline[user.UserName] = struct{}{}
	u.usersOnlineMutex.Unlock()
	return nil
}

func (u *UserManager) onUserLeft(msg any) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var user UserLeftMsg
	if err := json.Unmarshal(rawMsg, &user); err != nil {
		return err
	}
	u.usersOnlineMutex.Lock()
	delete(u.usersOnline, user.UserName)
	u.usersOnlineMutex.Unlock()
	return nil
}

func (u *UserManager) onUsersOnline(msg any) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var users UsersOnline
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
	var res []string
	u.usersOnlineMutex.Lock()
	for name := range u.usersOnline {
		res = append(res, name)
	}
	u.usersOnlineMutex.Unlock()
	return res
}

func (u *UserManager) notifyServerJoined() error {
	msg := &Msg{
		Type:   "server_joined",
		SentAt: time.Now(),
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) NotifyUserJoined(userName string) error {
	u.localUsersMutex.Lock()
	u.localUsers[userName] = struct{}{}
	u.localUsersMutex.Unlock()

	msg := &Msg{
		Type:   "user_joined",
		SentAt: time.Now(),
		Msg: &UserJoinedMsg{
			UserName: userName,
		},
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) NotifyUserLeft(userName string) error {
	u.localUsersMutex.Lock()
	u.localUsers[userName] = struct{}{}
	u.localUsersMutex.Unlock()

	msg := &Msg{
		Type:   "user_left",
		SentAt: time.Now(),
		Msg: &UserLeftMsg{
			UserName: userName,
		},
	}
	return u.pulbishMsg(msg)
}

func (u *UserManager) pulbishMsg(msg *Msg) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	u.users.Publish(rawMsg)
	return nil
}
