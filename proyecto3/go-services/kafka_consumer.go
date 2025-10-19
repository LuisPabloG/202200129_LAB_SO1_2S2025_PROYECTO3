package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader   *kafka.Reader
	redisDb  *redis.Client
}

func NewKafkaConsumer(kafkaAddr, redisAddr string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		Topic:   "weather-tweets",
		GroupID: "weather-consumer-kafka",
	})

	redisDb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &KafkaConsumer{
		reader:   reader,
		redisDb:  redisDb,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kc.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error leyendo de Kafka: %v", err)
				continue
			}

			var weather WeatherMessage
			err = json.Unmarshal(msg.Value, &weather)
			if err != nil {
				log.Printf("Error decodificando mensaje: %v", err)
				continue
			}

			// Procesar el mensaje: incrementar contadores en Redis
			kc.processWeatherData(&weather)
		}
	}
}

func (kc *KafkaConsumer) processWeatherData(weather *WeatherMessage) {
	ctx := context.Background()

	// Mapear enumeración a string
	weatherStr := mapWeatherEnum(weather.Weather)
	municipalityStr := mapMunicipalityEnum(weather.Municipality)

	// Crear clave para almacenar contador por clima
	key := "weather:" + weatherStr

	// Incrementar contador
	err := kc.redisDb.Incr(ctx, key).Err()
	if err != nil {
		log.Printf("Error incrementando contador: %v", err)
	}

	// Almacenar datos adicionales
	dataKey := "weather:data:" + municipalityStr + ":" + weatherStr
	kc.redisDb.HSet(ctx, dataKey, "temperature", weather.Temperature)
	kc.redisDb.HSet(ctx, dataKey, "humidity", weather.Humidity)

	log.Printf("Datos de clima procesados: municipio=%s, clima=%s", municipalityStr, weatherStr)
}

func (kc *KafkaConsumer) Close() error {
	if kc.reader != nil {
		kc.reader.Close()
	}
	if kc.redisDb != nil {
		kc.redisDb.Close()
	}
	return nil
}

func mapWeatherEnum(weather int32) string {
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

func mapMunicipalityEnum(municipality int32) string {
	switch municipality {
	case 1:
		return "mixco"
	case 2:
		return "guatemala"
	case 3:
		return "amatitlan"
	case 4:
		return "chinautla"
	default:
		return "unknown"
	}
}

func main() {
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "kafka:9092"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "valkey:6379"
	}

	log.Printf("Iniciando consumidor de Kafka...")
	consumer := NewKafkaConsumer(kafkaAddr, redisAddr)
	defer consumer.Close()

	ctx := context.Background()
	consumer.Start(ctx)
}
