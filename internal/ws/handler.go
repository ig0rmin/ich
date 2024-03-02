package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/user"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
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

	log.Printf("New webscoket connection\n")

	client := &Client{
		Conn:    conn,
		Message: make(chan string),
	}

	go func() {
		client.Message <- fmt.Sprintf(`{"message": "%s joined the chat"}`, c.GetString(user.UserNameKey))
	}()

	go client.write()
	client.read()

	log.Printf("Websocket client left")
}
