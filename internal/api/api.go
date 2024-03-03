package api

import "time"

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
