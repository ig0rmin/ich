package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	*Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{s}
}

func (h *Handler) Route(root gin.IRouter) {
	root.POST("/createUser", h.CreateUser)
	root.POST("/login", h.Login)
	root.POST("/logout", h.Logout)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var u CreateUserReq
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.Service.CreateUser(c.Request.Context(), &u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Service.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("jwt", resp.JWT, 60*60*24, "/", "localhost", false, true)

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
