package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	"go-services/proto"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedWeatherTweetServiceServer
}

func (s *server) SendTweet(ctx context.Context, in *proto.WeatherTweetRequest) (*proto.WeatherTweetResponse, error) {
	log.Printf("Received tweet via gRPC: %+v", in)

	// Convert the protobuf message to JSON
	tweetJSON, err := json.Marshal(in)
	if err != nil {
		log.Printf("Failed to marshal tweet to JSON: %v", err)
		return &proto.WeatherTweetResponse{Status: "Failed to process tweet"}, err
	}

	// Send to Kafka
	go sendToKafka(tweetJSON)

	// Send to RabbitMQ
	go sendToRabbitMQ(tweetJSON)

	return &proto.WeatherTweetResponse{Status: "Tweet received and is being processed"}, nil
}

func sendToKafka(message []byte) {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092" // Default for local/docker-compose
	}
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		log.Printf("Failed to create Kafka producer: %s", err)
		return
	}
	defer p.Close()

	topic := "weather-tweets"
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	// Wait for message deliveries
	p.Flush(15 * 1000)
	log.Println("Sent message to Kafka")
}

func sendToRabbitMQ(message []byte) {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/" // Default for local/docker-compose
	}
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %s", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weather-tweets", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %s", err)
		return
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		return
	}
	log.Println("Sent message to RabbitMQ")
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterWeatherTweetServiceServer(s, &server{})
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
