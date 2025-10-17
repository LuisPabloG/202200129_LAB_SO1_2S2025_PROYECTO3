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

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
)

// WeatherTweet es la estructura que se recibe de RabbitMQ
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
			Name: "rabbit_consumer_messages_processed_total",
			Help: "Total number of messages processed from RabbitMQ.",
		},
		[]string{"status", "municipality", "weather"},
	)

	processingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rabbit_consumer_processing_latency_seconds",
			Help:    "Time taken to process a message from RabbitMQ.",
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

// Consumer representa el consumidor de RabbitMQ
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	client  *redis.Client
	done    chan bool
	wg      sync.WaitGroup
}

// NewConsumer crea un nuevo consumidor
func NewConsumer(rabbitURL string, queueName string, redisAddr string) (*Consumer, error) {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}

	// Crear canal
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
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
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		client:  client,
		done:    make(chan bool),
	}, nil
}

// Start inicia el consumo de mensajes
func (c *Consumer) Start(queueName string) error {
	// Declarar la cola
	q, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	// Configurar QoS
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	// Consumir mensajes
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	// Procesar mensajes en una goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Canal de mensajes cerrado")
					return
				}
				c.processMessage(&msg)
				msg.Ack(false) // Confirmar procesamiento
			}
		}
	}()

	return nil
}

// processMessage procesa un mensaje de RabbitMQ
func (c *Consumer) processMessage(msg *amqp.Delivery) {
	timer := prometheus.NewTimer(processingLatency.WithLabelValues("success"))
	defer timer.ObserveDuration()

	var tweet WeatherTweet
	if err := json.Unmarshal(msg.Body, &tweet); err != nil {
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
	if err := c.channel.Close(); err != nil {
		log.Printf("Error al cerrar canal: %v", err)
	}
	if err := c.conn.Close(); err != nil {
		log.Printf("Error al cerrar conexión: %v", err)
	}
	if err := c.client.Close(); err != nil {
		log.Printf("Error al cerrar cliente Redis: %v", err)
	}
}

func main() {
	// Configurar variables de entorno
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	queueName := os.Getenv("RABBITMQ_QUEUE")
	if queueName == "" {
		queueName = "weather-tweets"
	}

	redisAddr := os.Getenv("VALKEY_ADDR")
	if redisAddr == "" {
		redisAddr = "valkey:6379"
	}

	// Crear consumidor
	consumer, err := NewConsumer(rabbitURL, queueName, redisAddr)
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
		log.Println("Iniciando servidor de métricas en :9093")
		if err := http.ListenAndServe(":9093", nil); err != nil {
			log.Fatalf("Error al iniciar servidor de métricas: %v", err)
		}
	}()

	// Iniciar consumo de mensajes
	err = consumer.Start(queueName)
	if err != nil {
		log.Fatalf("Error al iniciar consumo: %v", err)
	}
	log.Printf("Consumidor iniciado. Escuchando en cola '%s' de %s", queueName, rabbitURL)

	// Esperar señal de interrupción
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	log.Println("Apagando consumidor...")
	consumer.Stop()
	log.Println("Consumidor cerrado correctamente")
}
