package ws

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/kafka"
)

type Client struct {
	conn     *websocket.Conn
	messages *kafka.Kafka
	publish  chan []byte
}

func NewClient(conn *websocket.Conn, messages *kafka.Kafka) (*Client, error) {
	c := &Client{
		conn:     conn,
		messages: messages,
		publish:  make(chan []byte),
	}
	messages.Subscribe(c)
	return c, nil
}

func (c *Client) Receive(data []byte) error {
	c.publish <- data
	return nil
}

func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()

	for {
		msg, ok := <-c.publish
		if !ok {
			return
		}
		if err := c.conn.WriteJSON(gin.H{"message": string(msg)}); err != nil {
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
}
