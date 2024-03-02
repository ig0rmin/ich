package kafka

type Config struct {
	BootstrapServers []string `env:"ICH_KAFKA_BOOTSTRAP_SERVERS, delimiter=;, required"`
}
