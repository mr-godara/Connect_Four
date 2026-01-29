package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// InitKafka initializes the Kafka producer
func InitKafka() error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Println("KAFKA_BROKERS not set, analytics will be disabled")
		return nil
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "game-analytics"
	}

	writer = &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Printf("Kafka initialized: brokers=%s, topic=%s", brokers, topic)
	return nil
}

// SendAnalytics sends an analytics event to Kafka
func SendAnalytics(eventType string, data interface{}) {
	if writer == nil {
		return // Kafka not initialized
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling analytics event: %v", err)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(eventType),
			Value: eventJSON,
		})

		if err != nil {
			log.Printf("Kafka error (event: %s): %v", eventType, err)
		}
	}()
}

// Close closes the Kafka writer
func Close() {
	if writer != nil {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing Kafka writer: %v", err)
		}
	}
}
