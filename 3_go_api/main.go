package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/LuisPabloG/202200129_LAB_SO1_2S2025_PROYECTO3/3_go_api/proto"
)

// Definición de métricas Prometheus
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_api_requests_total",
			Help: "Total number of requests to the Go API.",
		},
		[]string{"status", "municipality", "weather"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "go_api_request_duration_seconds",
			Help:    "Duration of requests to the Go API.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

func init() {
	// Registrar métricas con Prometheus
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

// WeatherTweet representa la estructura de un tweet del clima
type WeatherTweet struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

// Mapa para convertir de string a enum de Municipio
var municipalityMap = map[string]pb.Municipalities{
	"mixco":     pb.Municipalities_mixco,
	"guatemala": pb.Municipalities_guatemala,
	"amatitlan": pb.Municipalities_amatitlan,
	"chinautla": pb.Municipalities_chinautla,
}

// Mapa para convertir de string a enum de Clima
var weatherMap = map[string]pb.Weathers{
	"sunny":  pb.Weathers_sunny,
	"cloudy": pb.Weathers_cloudy,
	"rainy":  pb.Weathers_rainy,
	"foggy":  pb.Weathers_foggy,
}

func main() {
	// Configuración del router Gin
	r := gin.Default()

	// Endpoint de métricas de Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Endpoint de estado para health checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Go API funcionando correctamente - 202200129",
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.POST("/tweet", handleTweet)
	}

	// Servidor HTTP
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Iniciar servidor en una goroutine
	go func() {
		log.Println("Iniciando servidor en :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar servidor: %s", err)
		}
	}()

	// Esperar señal de interrupción
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Apagando servidor...")

	// Contexto con timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Error al cerrar el servidor:", err)
	}
	log.Println("Servidor cerrado correctamente")
}

func handleTweet(c *gin.Context) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("/api/v1/tweet"))
	defer timer.ObserveDuration()

	var tweet WeatherTweet
	if err := c.BindJSON(&tweet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		requestsTotal.WithLabelValues("error", "unknown", "unknown").Inc()
		return
	}

	// Convertir a los enums del protobuf
	municipality, ok := municipalityMap[tweet.Municipality]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Municipio inválido"})
		requestsTotal.WithLabelValues("error", tweet.Municipality, tweet.Weather).Inc()
		return
	}

	weather, ok := weatherMap[tweet.Weather]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Clima inválido"})
		requestsTotal.WithLabelValues("error", tweet.Municipality, tweet.Weather).Inc()
		return
	}

	// Enviar a servicios gRPC en paralelo
	errChan := make(chan error, 2)

	// Enviar a servicio Kafka
	go func() {
		kafkaServiceAddr := os.Getenv("KAFKA_GRPC_SERVICE_ADDR")
		if kafkaServiceAddr == "" {
			kafkaServiceAddr = "kafka-writer:50051"
		}

		err := sendToGRPCService(kafkaServiceAddr, &pb.WeatherTweetRequest{
			Municipality: municipality,
			Temperature:  tweet.Temperature,
			Humidity:     tweet.Humidity,
			Weather:      weather,
		})
		errChan <- err
	}()

	// Enviar a servicio RabbitMQ
	go func() {
		rabbitServiceAddr := os.Getenv("RABBIT_GRPC_SERVICE_ADDR")
		if rabbitServiceAddr == "" {
			rabbitServiceAddr = "rabbit-writer:50052"
		}

		err := sendToGRPCService(rabbitServiceAddr, &pb.WeatherTweetRequest{
			Municipality: municipality,
			Temperature:  tweet.Temperature,
			Humidity:     tweet.Humidity,
			Weather:      weather,
		})
		errChan <- err
	}()

	// Esperar respuestas
	var errors []string
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"errors": errors,
		})
		requestsTotal.WithLabelValues("error", tweet.Municipality, tweet.Weather).Inc()
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Tweet procesado correctamente",
		})
		requestsTotal.WithLabelValues("success", tweet.Municipality, tweet.Weather).Inc()
	}
}

func sendToGRPCService(serviceAddr string, request *pb.WeatherTweetRequest) error {
	// Establecer conexión gRPC
	conn, err := grpc.Dial(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Error al conectar con servicio gRPC %s: %v", serviceAddr, err)
		return err
	}
	defer conn.Close()

	// Crear cliente gRPC
	client := pb.NewWeatherTweetServiceClient(conn)

	// Timeout para la llamada
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Enviar request
	_, err = client.SendTweet(ctx, request)
	if err != nil {
		log.Printf("Error al enviar tweet a servicio gRPC %s: %v", serviceAddr, err)
		return err
	}

	return nil
}
