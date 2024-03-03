package ws

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/kafka"
)

type Client struct {
	Conn     *websocket.Conn
	messages *kafka.Kafka
	publish  chan []byte
}

func NewClient(conn *websocket.Conn, messages *kafka.Kafka) (*Client, error) {
	c := &Client{
		Conn:     conn,
		messages: messages,
		publish:  make(chan []byte),
	}
	//messages.Subscribe(c)
	return c, nil
}

func (c *Client) Receive(data []byte) error {
	c.publish <- data
	return nil
}

func (c *Client) write() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		msg, ok := <-c.publish
		if !ok {
			return
		}
		if err := c.Conn.WriteJSON(gin.H{"message": msg}); err != nil {
			break
		}
	}
}

func (c *Client) read() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("Websockets received message: %v\n", string(msg))

		//c.messages.Publish(msg)
	}
}

func (c *Client) Close() {
	c.messages.Unsubscribe(c)
}
