package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ig0rmin/ich/internal/messages"
	"github.com/ig0rmin/ich/internal/user"
	"github.com/ig0rmin/ich/internal/users"
)

type Handler struct {
	msg     *messages.Messages
	userMgr *users.UserManager
}

func NewHandler(userMgr *users.UserManager, msg *messages.Messages) *Handler {
	return &Handler{
		msg:     msg,
		userMgr: userMgr,
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

	userName := c.GetString(user.UserNameKey)
	h.userMgr.NotifyUserJoined(userName)
	defer h.userMgr.NotifyUserLeft(userName)

	client, err := NewClient(conn, userName, h.userMgr, h.msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Client"})
		return
	}
	defer client.Close()

	client.Init()

	go client.write()
	client.read()

	log.Printf("Websocket client left")
}
