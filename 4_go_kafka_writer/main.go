package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/LuisPabloG/202200129_LAB_SO1_2S2025_PROYECTO3/4_go_kafka_writer/proto"
)

// Definición de métricas Prometheus
var (
	messagesPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_writer_messages_published_total",
			Help: "Total number of messages published to Kafka.",
		},
		[]string{"status", "municipality", "weather"},
	)

	messageLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_writer_message_latency_seconds",
			Help:    "Time taken to publish a message to Kafka.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

func init() {
	// Registrar métricas con Prometheus
	prometheus.MustRegister(messagesPublished)
	prometheus.MustRegister(messageLatency)
}

// Server es la implementación del servicio gRPC
type Server struct {
	pb.UnimplementedWeatherTweetServiceServer
	producer sarama.SyncProducer
	topic    string
}

// Tweet es la estructura que se enviará a Kafka
type Tweet struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
	Timestamp    string `json:"timestamp"`
}

// SendTweet implementa la función definida en el protobuf
func (s *Server) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	timer := prometheus.NewTimer(messageLatency.WithLabelValues("success"))
	defer timer.ObserveDuration()

	// Convertir el municipio y clima de enums a strings
	municipality := req.Municipality.String()
	weather := req.Weather.String()

	// Crear un tweet con timestamp
	tweet := Tweet{
		Municipality: municipality,
		Temperature:  req.Temperature,
		Humidity:     req.Humidity,
		Weather:      weather,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	// Convertir a JSON
	jsonData, err := json.Marshal(tweet)
	if err != nil {
		log.Printf("Error al serializar tweet: %v", err)
		messagesPublished.WithLabelValues("error", municipality, weather).Inc()
		return &pb.WeatherTweetResponse{Status: "error"}, err
	}

	// Crear mensaje para Kafka
	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(jsonData),
	}

	// Publicar en Kafka
	_, _, err = s.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error al enviar mensaje a Kafka: %v", err)
		messagesPublished.WithLabelValues("error", municipality, weather).Inc()
		return &pb.WeatherTweetResponse{Status: "error"}, err
	}

	log.Printf("Mensaje publicado en Kafka: %s", string(jsonData))
	messagesPublished.WithLabelValues("success", municipality, weather).Inc()

	return &pb.WeatherTweetResponse{Status: "success"}, nil
}

// HealthServer implementa el servicio de health check de gRPC
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

// Check implementa la función de health check
func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch implementa la función watch del health check
func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func main() {
	// Configurar variables de entorno
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "weather-tweets"
	}

	// Configurar Sarama (cliente Kafka)
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// Conectar con Kafka
	producer, err := sarama.NewSyncProducer([]string{kafkaBrokers}, config)
	if err != nil {
		log.Fatalf("Error al crear productor Kafka: %v", err)
	}
	defer producer.Close()

	// Crear servidor gRPC
	server := &Server{
		producer: producer,
		topic:    kafkaTopic,
	}

	// Iniciar servidor HTTP para métricas de Prometheus
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Iniciando servidor de métricas en :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Error al iniciar servidor de métricas: %v", err)
		}
	}()

	// Iniciar servidor gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al escuchar en puerto: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, server)
	grpc_health_v1.RegisterHealthServer(s, &HealthServer{})
	reflection.Register(s)

	// Iniciar servidor en una goroutine
	go func() {
		log.Println("Iniciando servidor gRPC en :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Error al iniciar servidor gRPC: %v", err)
		}
	}()

	// Esperar señal de interrupción
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Apagando servidor...")
	s.GracefulStop()
	log.Println("Servidor cerrado correctamente")
}
