package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/kafka"
)

type Handler struct {
	messages *kafka.Kafka
}

func NewHandler(messages *kafka.Kafka) *Handler {
	return &Handler{
		messages: messages,
	}
}

func (h *Handler) Route(root gin.IRouter) {
	root.GET("/join", h.Join)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) Join(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("New webscoket connection")

	client, err := NewClient(conn, h.messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Client"})
		return
	}
	defer client.Close()

	go client.write()
	client.read()

	log.Printf("Websocket client left")
}
