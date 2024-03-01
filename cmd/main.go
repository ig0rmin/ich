package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type DBConfig struct {
	DBHost     string `env:"ICH_DB_HOST, required"`
	DBPort     string `env:"ICH_DB_PORT, required"`
	DBUser     string `env:"ICH_DB_USER, required"`
	DBPassword string `env:"ICH_DB_PASSWORD, required"`
	DBName     string `end:"ICH_DB_NAME, required"`
}

type KafkaConfig struct {
	BootstrapServers []string `env:"ICH_KAFKA_BOOTSTRAP_SERVERS, delimiter=;, required"`
}

type Config struct {
	// Port to listen
	Port string `env:"ICH_PORT, default=8080"`

	DB    DBConfig
	Kafka KafkaConfig
}

type Server struct {
	Cfg Config
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
}
