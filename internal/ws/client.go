package ws

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	Message chan string
}

// FIXME: this is echo client to test websocket connection

func (c *Client) write() {
	defer func() {
		c.Conn.Close()
	}()

	if err := c.Conn.WriteJSON(gin.H{"message": "hello to the lobby"}); err != nil {
		return
	}

	for {
		msg, ok := <-c.Message
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

		log.Printf("Received message: %v\n", string(msg))
		go func() {
			c.Message <- string(msg)
		}()
	}
}
