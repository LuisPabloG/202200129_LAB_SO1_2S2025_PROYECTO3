package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

var ctx = context.Background()

type Tweet struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
}

func main() {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})

	conn, err := amqp.Dial(rabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weather-tweets", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received from RabbitMQ: %s", d.Body)
			var tweet Tweet
			if err := json.Unmarshal(d.Body, &tweet); err != nil {
				log.Printf("Error unmarshalling tweet: %v", err)
				continue
			}

			// Store in Valkey
			weatherCondition := getWeatherCondition(tweet.Weather)
			err = rdb.Incr(ctx, fmt.Sprintf("weather:%s", weatherCondition)).Err()
			if err != nil {
				log.Printf("Failed to increment weather condition count in Valkey: %v", err)
			} else {
				log.Printf("Incremented count for weather: %s", weatherCondition)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
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
