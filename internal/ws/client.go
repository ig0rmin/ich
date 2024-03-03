package ws

import (
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/api"
	"github.com/ig0rmin/ich/internal/kafka"
	"github.com/ig0rmin/ich/internal/users"
)

type Client struct {
	conn *websocket.Conn
	//TODO: Replace by MessageManager
	messages *kafka.Kafka
	userMgr  *users.UserManager
	publish  chan any
}

func NewClient(conn *websocket.Conn, userMgr *users.UserManager, messages *kafka.Kafka) (*Client, error) {
	c := &Client{
		conn:     conn,
		messages: messages,
		userMgr:  userMgr,
		publish:  make(chan any),
	}
	messages.Subscribe(c)
	c.userMgr.Subscribe(c)
	return c, nil
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

// TODO: Typed data
func (c *Client) Receive(data []byte) error {
	c.publish <- data
	return nil
}

func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()

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
	defer func() {
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("Websockets received message: %v", strings.TrimSuffix(string(msg), "\n"))

		c.messages.Publish(msg)
	}
}

func (c *Client) Close() {
	c.messages.Unsubscribe(c)
	c.userMgr.Unsubscribe(c)
}
