package ws

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/messages"
	"github.com/ig0rmin/ich/internal/users"
)

type Client struct {
	userName string
	conn     *websocket.Conn
	messages *messages.Messages
	userMgr  *users.UserManager
	publish  chan any
}

func NewClient(conn *websocket.Conn, userName string, userMgr *users.UserManager, msg *messages.Messages) (*Client, error) {
	c := &Client{
		userName: userName,
		conn:     conn,
		messages: msg,
		userMgr:  userMgr,
		publish:  make(chan any),
	}
	return c, nil
}

func (c *Client) Init() {
	c.messages.Subscribe(c)
	c.userMgr.Subscribe(c)
}

func (c *Client) Close() {
	c.messages.Unsubscribe(c)
	c.userMgr.Unsubscribe(c)
}

func (c *Client) ReceiveChatMessage(sentAt time.Time, chatMsg *api.ChatMessage) {
	msg := &api.Msg{
		Type:   api.TypeChatMessage,
		SentAt: time.Now(),
		Msg:    chatMsg,
	}
	c.publish <- msg
}

func (c *Client) ReceiveUserJoined(user *api.UserJoinedMsg) {
	msg := &api.Msg{
		Type:   api.TypeUserJoined,
		SentAt: time.Now(),
		Msg:    user,
	}
	c.publish <- msg
}

func (c *Client) ReceiveUserLeft(user *api.UserLeftMsg) {
	msg := &api.Msg{
		Type:   api.TypeUserLeft,
		SentAt: time.Now(),
		Msg:    user,
	}
	c.publish <- msg
}

func (c *Client) write() {
	defer c.conn.Close()

	// Send the list of users online as the first message to the new client
	if err := c.conn.WriteJSON(c.usersOnline()); err != nil {
		return
	}

	for {
		msg, ok := <-c.publish
		if !ok {
			return
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			break
		}
	}
}

func (c *Client) read() {
	defer c.conn.Close()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("Websockets received message: %v", strings.TrimSuffix(string(msg), "\n"))
		c.processChatMessage(msg)
	}
}

func (c *Client) processChatMessage(data []byte) {
	chatMsg := &api.ChatMessage{}
	msg := &api.Msg{
		Msg: chatMsg,
	}
	if err := json.Unmarshal(data, msg); err != nil {
		log.Printf("Can't parse JSON")
		return
	}
	if msg.Type != api.TypeChatMessage {
		log.Printf("Unsupported message type: %v", msg.Type)
		return
	}
	// Prevent spoofing user name
	chatMsg.From = c.userName

	c.messages.PostChatMessage(chatMsg)
}

func (c *Client) usersOnline() *api.Msg {
	msg := &api.Msg{
		Type:   api.TypeUsersOnline,
		SentAt: time.Now(),
		Msg: &api.UsersOnline{
			List: c.userMgr.GetUsersOnline(),
		},
	}
	return msg
}
