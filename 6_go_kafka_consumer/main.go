package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WeatherTweet es la estructura que se recibe de Kafka
type WeatherTweet struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
	Timestamp    string `json:"timestamp"`
}

// Definición de métricas Prometheus
var (
	messagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumer_messages_processed_total",
			Help: "Total number of messages processed from Kafka.",
		},
		[]string{"status", "municipality", "weather"},
	)

	processingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_processing_latency_seconds",
			Help:    "Time taken to process a message from Kafka.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

func init() {
	// Registrar métricas con Prometheus
	prometheus.MustRegister(messagesProcessed)
	prometheus.MustRegister(processingLatency)
}

// Consumer representa el consumidor de Kafka
type Consumer struct {
	consumer sarama.Consumer
	client   *redis.Client
	done     chan bool
	wg       sync.WaitGroup
}

// NewConsumer crea un nuevo consumidor
func NewConsumer(brokers []string, topic string, redisAddr string) (*Consumer, error) {
	// Configuración de Kafka
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Crear consumidor de Kafka
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	// Configurar cliente de Valkey (compatible con Redis)
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	// Verificar conexión con Valkey
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		client:   client,
		done:     make(chan bool),
	}, nil
}

// Start inicia el consumo de mensajes
func (c *Consumer) Start(topic string) {
	// Obtener particiones disponibles
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		log.Fatalf("Error al obtener particiones: %v", err)
	}

	// Consumir de cada partición en una goroutine separada
	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error al consumir partición %d: %v", partition, err)
			continue
		}

		c.wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer c.wg.Done()
			for {
				select {
				case msg := <-pc.Messages():
					c.processMessage(msg)
				case err := <-pc.Errors():
					log.Printf("Error al consumir mensaje: %v", err)
				case <-c.done:
					return
				}
			}
		}(pc)
	}
}

// processMessage procesa un mensaje de Kafka
func (c *Consumer) processMessage(msg *sarama.ConsumerMessage) {
	timer := prometheus.NewTimer(processingLatency.WithLabelValues("success"))
	defer timer.ObserveDuration()

	var tweet WeatherTweet
	if err := json.Unmarshal(msg.Value, &tweet); err != nil {
		log.Printf("Error al deserializar mensaje: %v", err)
		messagesProcessed.WithLabelValues("error", "unknown", "unknown").Inc()
		return
	}

	// Limpiar los nombres de municipio y clima para usar como llaves
	municipality := strings.TrimPrefix(tweet.Municipality, "municipalities_")
	weather := strings.TrimPrefix(tweet.Weather, "weathers_")

	ctx := context.Background()

	// Almacenar datos en Valkey
	// 1. Incrementar contador total
	_, err := c.client.Incr(ctx, "total_tweets").Result()
	if err != nil {
		log.Printf("Error al incrementar total_tweets: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 2. Incrementar contador por municipio
	_, err = c.client.Incr(ctx, "municipality:"+municipality).Result()
	if err != nil {
		log.Printf("Error al incrementar contador de municipio: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 3. Incrementar contador por clima
	_, err = c.client.Incr(ctx, "weather:"+weather).Result()
	if err != nil {
		log.Printf("Error al incrementar contador de clima: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 4. Incrementar contador por municipio y clima
	_, err = c.client.Incr(ctx, "municipality:"+municipality+":weather:"+weather).Result()
	if err != nil {
		log.Printf("Error al incrementar contador municipio+clima: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 5. Almacenar temperatura promedio por municipio
	// Primero obtenemos la temperatura actual y el contador
	tempSum, err := c.client.Get(ctx, "municipality:"+municipality+":temp_sum").Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error al obtener suma de temperatura: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}
	tempSumVal := 0
	if err != redis.Nil {
		tempSumVal, _ = strconv.Atoi(tempSum)
	}
	tempSumVal += int(tweet.Temperature)

	// Almacenar la suma actualizada
	_, err = c.client.Set(ctx, "municipality:"+municipality+":temp_sum", tempSumVal, 0).Result()
	if err != nil {
		log.Printf("Error al almacenar suma de temperatura: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 6. Obtener el contador para calcular el promedio
	count, err := c.client.Get(ctx, "municipality:"+municipality).Result()
	if err != nil {
		log.Printf("Error al obtener contador de municipio: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}
	countVal, _ := strconv.Atoi(count)

	// Calcular y almacenar el promedio
	if countVal > 0 {
		avgTemp := float64(tempSumVal) / float64(countVal)
		_, err = c.client.Set(ctx, "municipality:"+municipality+":avg_temp", avgTemp, 0).Result()
		if err != nil {
			log.Printf("Error al almacenar temperatura promedio: %v", err)
			messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
			return
		}
	}

	// 7. Almacenar humedad promedio por municipio (similar a temperatura)
	humSum, err := c.client.Get(ctx, "municipality:"+municipality+":hum_sum").Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error al obtener suma de humedad: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}
	humSumVal := 0
	if err != redis.Nil {
		humSumVal, _ = strconv.Atoi(humSum)
	}
	humSumVal += int(tweet.Humidity)

	_, err = c.client.Set(ctx, "municipality:"+municipality+":hum_sum", humSumVal, 0).Result()
	if err != nil {
		log.Printf("Error al almacenar suma de humedad: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// Calcular y almacenar el promedio de humedad
	if countVal > 0 {
		avgHum := float64(humSumVal) / float64(countVal)
		_, err = c.client.Set(ctx, "municipality:"+municipality+":avg_hum", avgHum, 0).Result()
		if err != nil {
			log.Printf("Error al almacenar humedad promedio: %v", err)
			messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
			return
		}
	}

	// 8. Almacenar timestamp del último tweet
	_, err = c.client.Set(ctx, "last_updated", time.Now().Format(time.RFC3339), 0).Result()
	if err != nil {
		log.Printf("Error al almacenar timestamp: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// 9. Guardar el tweet completo en una lista para referencia
	tweetJSON, _ := json.Marshal(tweet)
	_, err = c.client.LPush(ctx, "tweets", tweetJSON).Result()
	if err != nil {
		log.Printf("Error al almacenar tweet en lista: %v", err)
		messagesProcessed.WithLabelValues("error", municipality, weather).Inc()
		return
	}

	// Opcional: Limitar el tamaño de la lista
	_, err = c.client.LTrim(ctx, "tweets", 0, 999).Result() // Mantener solo los últimos 1000 tweets
	if err != nil {
		log.Printf("Error al limitar lista de tweets: %v", err)
	}

	// Establecer TTL para evitar saturación (opcional)
	// Esto configura un tiempo de expiración de 24 horas para los datos
	ttl := 24 * time.Hour
	if ttlEnv := os.Getenv("DATA_TTL_HOURS"); ttlEnv != "" {
		if ttlHours, err := strconv.Atoi(ttlEnv); err == nil {
			ttl = time.Duration(ttlHours) * time.Hour
		}
	}

	if ttl > 0 {
		_, err = c.client.Expire(ctx, "tweets", ttl).Result()
		if err != nil {
			log.Printf("Error al configurar TTL para tweets: %v", err)
		}
	}

	log.Printf("Mensaje procesado correctamente: %s - %s (Temp: %d, Hum: %d)",
		municipality, weather, tweet.Temperature, tweet.Humidity)
	messagesProcessed.WithLabelValues("success", municipality, weather).Inc()
}

// Stop detiene el consumo de mensajes
func (c *Consumer) Stop() {
	close(c.done)
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		log.Printf("Error al cerrar consumidor: %v", err)
	}
	if err := c.client.Close(); err != nil {
		log.Printf("Error al cerrar cliente Redis: %v", err)
	}
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

	redisAddr := os.Getenv("VALKEY_ADDR")
	if redisAddr == "" {
		redisAddr = "valkey:6379"
	}

	// Crear consumidor
	consumer, err := NewConsumer(strings.Split(kafkaBrokers, ","), kafkaTopic, redisAddr)
	if err != nil {
		log.Fatalf("Error al crear consumidor: %v", err)
	}

	// Iniciar servidor HTTP para métricas de Prometheus
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	go func() {
		log.Println("Iniciando servidor de métricas en :9092")
		if err := http.ListenAndServe(":9092", nil); err != nil {
			log.Fatalf("Error al iniciar servidor de métricas: %v", err)
		}
	}()

	// Iniciar consumo de mensajes
	consumer.Start(kafkaTopic)
	log.Printf("Consumidor iniciado. Escuchando en topic '%s' de %s", kafkaTopic, kafkaBrokers)

	// Esperar señal de interrupción
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	log.Println("Apagando consumidor...")
	consumer.Stop()
	log.Println("Consumidor cerrado correctamente")
}
