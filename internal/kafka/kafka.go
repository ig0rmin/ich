package kafka

import (
	"context"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Config struct {
	BootstrapServers []string `env:"ICH_KAFKA_BOOTSTRAP_SERVERS, delimiter=;, required"`
}

type Kafka struct {
	topic             string
	producer          sarama.SyncProducer
	consumer          sarama.Consumer
	partitionConsumer sarama.PartitionConsumer
	publish           chan []byte

	receivers      map[Receiver]struct{}
	receiversMutex sync.Mutex

	wg *sync.WaitGroup
}

type Receiver interface {
	Receive([]byte) error
}

func NewKafka(cfg Config, topic string) (*Kafka, error) {
	producer, err := connectProducer(cfg.BootstrapServers)
	if err != nil {
		return nil, err
	}

	consumer, err := connectConsumer(cfg.BootstrapServers)
	if err != nil {
		producer.Close()
		return nil, err
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	return &Kafka{
		topic:             topic,
		producer:          producer,
		consumer:          consumer,
		partitionConsumer: partitionConsumer,
		publish:           make(chan []byte),
		receivers:         make(map[Receiver]struct{}),
		wg:                &sync.WaitGroup{},
	}, nil
}

func (k *Kafka) Subscribe(r Receiver) {
	k.receiversMutex.Lock()
	k.receivers[r] = struct{}{}
	k.receiversMutex.Unlock()
}

func (k *Kafka) Unsubscribe(r Receiver) {
	k.receiversMutex.Lock()
	delete(k.receivers, r)
	k.receiversMutex.Unlock()
}

func (k *Kafka) Publish(data []byte) {
	k.publish <- data
}

func (k *Kafka) Run(ctx context.Context) {
	go k.consume(ctx)
	k.produce(ctx)
}

func (k *Kafka) Wait() {
	k.wg.Wait()
}

func (k *Kafka) Close() {
	k.producer.Close()
	k.partitionConsumer.Close()
	k.consumer.Close()
}

func (k *Kafka) consume(ctx context.Context) {
	log.Printf("Start Kafka consumer loop for the topic %v", k.topic)
	k.wg.Add(1)
Loop:
	for {
		select {
		case err := <-k.partitionConsumer.Errors():
			if err == nil {
				log.Fatalf("Received nil message from Kafka, probably closed connection")
			}
			log.Println(err.Error())
		case msg := <-k.partitionConsumer.Messages():
			if msg == nil {
				log.Fatalf("Received nil message from Kafka, probably closed connection")
			}
			log.Printf("Kafka receidved message %v from the topic %v", string(msg.Value), k.topic)
			k.receiversMutex.Lock()
			for r := range k.receivers {
				r.Receive(msg.Value)
			}
			k.receiversMutex.Unlock()
		case <-ctx.Done():
			break Loop
		}
	}
	k.wg.Done()
	log.Printf("Done Kafka consumer loop for the topic %v", k.topic)

}

func (k *Kafka) produce(ctx context.Context) {
	log.Printf("Start Kafka producer loop for the topic %v", k.topic)
	k.wg.Add(1)
Loop:
	for {
		select {
		case data := <-k.publish:
			msg := &sarama.ProducerMessage{
				Topic: k.topic,
				Value: sarama.ByteEncoder(data),
			}
			_, _, err := k.producer.SendMessage(msg)
			if err != nil {
				log.Println(err.Error())
			}
		case <-ctx.Done():
			break Loop
		}
	}
	log.Printf("Done Kafka producer loop for the topic %v", k.topic)
	k.wg.Done()
}

func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
