package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ig0rmin/ich/internal/db"
	"github.com/ig0rmin/ich/internal/kafka"
	"github.com/ig0rmin/ich/internal/user"
	"github.com/ig0rmin/ich/internal/ws"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Port to listen
	Port         string `env:"ICH_PORT, default=8080"`
	ServerSecret string `env:"ICH_SERVER_SECRET, required"`

	DB    db.Config
	Kafka kafka.Config
}

func (c *Config) Load() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	if err := envconfig.Process(context.Background(), c); err != nil {
		return err
	}
	return nil
}

type Server struct {
	db     *sql.DB
	router *gin.Engine
	server *http.Server
}

func NewServer(cfg *Config) (*Server, error) {
	var err error
	s := &Server{}

	s.db, err = db.Connect(&cfg.DB)
	if err != nil {
		return nil, err
	}
	err = s.db.Ping()
	if err != nil {
		return nil, err
	}
	log.Println("Connection to DB established")

	if err := db.Migrate(&cfg.DB); err != nil {
		return nil, err
	}

	s.router = gin.Default()

	// Set up routes
	s.router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	user.NewHandler(
		user.NewService(user.NewRepository(s.db), cfg.ServerSecret),
	).Route(s.router)

	authenticated := s.router.Group("", authMiddleware(cfg.ServerSecret))
	authenticated.GET("/auth-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			user.UserNameKey: c.GetString(user.UserNameKey),
			user.UserIDKey:   c.GetString(user.UserIDKey),
		})
	})

	ws.NewHandler().Route(authenticated)

	s.server = &http.Server{
		Addr:    "0.0.0.0:" + cfg.Port,
		Handler: s.router,
	}

	return s, nil
}

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Close() {
	s.db.Close()
}
