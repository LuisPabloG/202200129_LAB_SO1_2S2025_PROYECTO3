package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"
	pb "weather-go-services/proto"
)

// Deployment 1: API REST + gRPC Client (este también actúa como servidor gRPC)

type WeatherServiceServer struct {
	pb.UnimplementedWeatherTweetServiceServer
	kafkaWriter  *KafkaWriter
	rabbitmqSender *RabbitMQSender
}

func (s *WeatherServiceServer) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	log.Printf("Recibido tweet: municipio=%d, temp=%d, humedad=%d, clima=%d", 
		req.Municipality, req.Temperature, req.Humidity, req.Weather)

	// Enviar a Kafka
	if s.kafkaWriter != nil {
		err := s.kafkaWriter.Send(req)
		if err != nil {
			log.Printf("Error enviando a Kafka: %v", err)
		}
	}

	// Enviar a RabbitMQ
	if s.rabbitmqSender != nil {
		err := s.rabbitmqSender.Send(req)
		if err != nil {
			log.Printf("Error enviando a RabbitMQ: %v", err)
		}
	}

	return &pb.WeatherTweetResponse{
		Status: "Tweet procesado exitosamente",
	}, nil
}

func main() {
	// Puerto para gRPC
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// Puerto para REST API
	restPort := os.Getenv("REST_PORT")
	if restPort == "" {
		restPort = "8081"
	}

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "kafka:9092"
	}

	rabbitmqAddr := os.Getenv("RABBITMQ_ADDR")
	if rabbitmqAddr == "" {
		rabbitmqAddr = "amqp://guest:guest@rabbitmq:5672/"
	}

	// Inicializar Kafka Writer
	kafkaWriter := NewKafkaWriter(kafkaAddr)
	defer kafkaWriter.Close()

	// Inicializar RabbitMQ Sender
	rabbitmqSender := NewRabbitMQSender(rabbitmqAddr)
	defer rabbitmqSender.Close()

	// Crear servidor gRPC
	server := &WeatherServiceServer{
		kafkaWriter: kafkaWriter,
		rabbitmqSender: rabbitmqSender,
	}

	// Iniciar servidor gRPC
	go startGRPCServer(server, grpcPort)

	// Iniciar servidor REST API
	go startRESTServer(restPort)

	// Mantener la aplicación activa
	select {}
}

func startGRPCServer(server *WeatherServiceServer, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("No se pudo escuchar en puerto %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(grpcServer, server)

	log.Printf("Servidor gRPC escuchando en puerto %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error en servidor gRPC: %v", err)
	}
}

func startRESTServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	log.Printf("Servidor REST escuchando en puerto %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Error en servidor REST: %v", err)
	}
}
