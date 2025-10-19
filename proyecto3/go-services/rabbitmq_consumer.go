package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	redisDb *redis.Client
}

func NewRabbitMQConsumer(amqpAddr, redisAddr string) *RabbitMQConsumer {
	conn, err := amqp.Dial(amqpAddr)
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}

	// Declarar exchange
	err = ch.ExchangeDeclare("weather-exchange", "direct", true, false, false, false, nil)
	if err != nil {
		log.Printf("Error declarando exchange: %v", err)
	}

	// Declarar queue
	q, err := ch.QueueDeclare("weather-tweets", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error declarando queue: %v", err)
	}

	// Bind queue a exchange
	err = ch.QueueBind(q.Name, "weather", "weather-exchange", false, nil)
	if err != nil {
		log.Printf("Error haciendo bind: %v", err)
	}

	redisDb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		redisDb: redisDb,
	}
}

func (rc *RabbitMQConsumer) Start(ctx context.Context) {
	msgs, err := rc.channel.Consume(
		"weather-tweets",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Error configurando consumer: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgs:
			if msg == nil {
				continue
			}

			var weather WeatherMessage
			err := json.Unmarshal(msg.Body, &weather)
			if err != nil {
				log.Printf("Error decodificando mensaje: %v", err)
				continue
			}

			rc.processWeatherData(&weather)
		}
	}
}

func (rc *RabbitMQConsumer) processWeatherData(weather *WeatherMessage) {
	ctx := context.Background()

	weatherStr := mapWeatherEnum(weather.Weather)
	municipalityStr := mapMunicipalityEnum(weather.Municipality)

	key := "weather:" + weatherStr

	err := rc.redisDb.Incr(ctx, key).Err()
	if err != nil {
		log.Printf("Error incrementando contador: %v", err)
	}

	dataKey := "weather:data:" + municipalityStr + ":" + weatherStr
	rc.redisDb.HSet(ctx, dataKey, "temperature", weather.Temperature)
	rc.redisDb.HSet(ctx, dataKey, "humidity", weather.Humidity)

	log.Printf("Datos de clima procesados: municipio=%s, clima=%s", municipalityStr, weatherStr)
}

func (rc *RabbitMQConsumer) Close() error {
	if rc.channel != nil {
		rc.channel.Close()
	}
	if rc.conn != nil {
		rc.conn.Close()
	}
	if rc.redisDb != nil {
		rc.redisDb.Close()
	}
	return nil
}

func main() {
	amqpAddr := os.Getenv("RABBITMQ_ADDR")
	if amqpAddr == "" {
		amqpAddr = "amqp://guest:guest@rabbitmq:5672/"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "valkey:6379"
	}

	log.Printf("Iniciando consumidor de RabbitMQ...")
	consumer := NewRabbitMQConsumer(amqpAddr, redisAddr)
	defer consumer.Close()

	ctx := context.Background()
	consumer.Start(ctx)
}
