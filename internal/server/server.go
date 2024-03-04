package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ig0rmin/ich/internal/db"
	"github.com/ig0rmin/ich/internal/kafka"
	"github.com/ig0rmin/ich/internal/messages"
	"github.com/ig0rmin/ich/internal/user"
	"github.com/ig0rmin/ich/internal/users"
	"github.com/ig0rmin/ich/internal/ws"
)

type Config struct {
	// Port to listen
	Port         string `env:"ICH_PORT, default=8080"`
	ServerSecret string `env:"ICH_SERVER_SECRET, required"`

	DB    db.Config
	Kafka kafka.Config
}

type Server struct {
	db       *sql.DB
	messages *kafka.Kafka
	users    *kafka.Kafka

	userMgr *users.UserManager
	msg     *messages.Messages

	server *http.Server
	router *gin.Engine
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

	s.messages, err = kafka.NewKafka(cfg.Kafka, "topic-messages")
	if err != nil {
		return nil, err
	}
	s.users, err = kafka.NewKafka(cfg.Kafka, "topic-users")
	if err != nil {
		return nil, err
	}

	s.userMgr, err = users.NewUserManager(s.users)
	if err != nil {
		return nil, err
	}

	s.msg, err = messages.NewMessages(s.messages)
	if err != nil {
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

	ws.NewHandler(s.userMgr, s.msg).Route(authenticated)

	s.server = &http.Server{
		Addr:    "0.0.0.0:" + cfg.Port,
		Handler: s.router,
	}

	return s, nil
}

func (s *Server) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.messages.Run(ctx)
	go s.users.Run(ctx)

	s.userMgr.Init()
	s.msg.Init()

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	s.waitForInterrupt()

	log.Println("Http server done")

	// Stop Kafka
	cancel()

	s.messages.Wait()
	s.users.Wait()

	log.Println("Kafka done")
}

func (s *Server) waitForInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	// Timeout for gracefull shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}
}

func (s *Server) Close() {
	s.db.Close()
	s.messages.Close()
	s.users.Close()
	s.msg.Close()
}
