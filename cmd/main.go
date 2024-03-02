package main

import (
	"log"

	"github.com/ig0rmin/ich/internal/server"
)

func main() {
	var cfg server.Config
	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("%v\n", cfg)

	s, err := server.NewServer(&cfg)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer s.Close()

	s.Run()
}
