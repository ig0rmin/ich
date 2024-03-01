package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ig0rmin/ich/db"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type KafkaConfig struct {
	BootstrapServers []string `env:"ICH_KAFKA_BOOTSTRAP_SERVERS, delimiter=;, required"`
}

type Config struct {
	// Port to listen
	Port string `env:"ICH_PORT, default=8080"`

	DB    db.Config
	Kafka KafkaConfig
}

type Server struct {
	db *sql.DB
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", cfg)

	dbConn, err := db.Connect(&cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connection to DB established")

	if err := db.Migrate(dbConn); err != nil {
		log.Fatalf("Failed to migrate DB")
	}
}
