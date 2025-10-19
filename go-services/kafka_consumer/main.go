package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Tweet struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
}

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          "weather-consumer-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	c.SubscribeTopics([]string{"weather-tweets"}, nil)
	log.Println("Kafka consumer started. Waiting for messages...")

	for {
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			log.Printf("Received from Kafka: %s", string(msg.Value))
			var tweet Tweet
			if err := json.Unmarshal(msg.Value, &tweet); err != nil {
				log.Printf("Error unmarshalling tweet: %v", err)
				continue
			}

			// Store in Valkey
			// Example: store total reports per weather condition
			weatherCondition := getWeatherCondition(tweet.Weather)
			err = rdb.Incr(ctx, fmt.Sprintf("weather:%s", weatherCondition)).Err()
			if err != nil {
				log.Printf("Failed to increment weather condition count in Valkey: %v", err)
			} else {
				log.Printf("Incremented count for weather: %s", weatherCondition)
			}

		} else if !err.(kafka.Error).IsTimeout() {
			// The client will automatically try to recover from all errors.
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}

func getWeatherCondition(weather int32) string {
	switch weather {
	case 1:
		return "sunny"
	case 2:
		return "cloudy"
	case 3:
		return "rainy"
	case 4:
		return "foggy"
	default:
		return "unknown"
	}
}
