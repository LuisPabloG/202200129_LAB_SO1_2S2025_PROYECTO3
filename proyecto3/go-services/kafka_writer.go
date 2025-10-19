package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	pb "weather-go-services/proto"
)

type KafkaWriter struct {
	writer *kafka.Writer
}

type WeatherMessage struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
	Timestamp    int64 `json:"timestamp"`
}

func NewKafkaWriter(addr string) *KafkaWriter {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{addr},
		Topic:    "weather-tweets",
		MaxBytes: 1e6,
	})
	return &KafkaWriter{writer: writer}
}

func (kw *KafkaWriter) Send(req *pb.WeatherTweetRequest) error {
	msg := WeatherMessage{
		Municipality: req.Municipality,
		Temperature:  req.Temperature,
		Humidity:     req.Humidity,
		Weather:      req.Weather,
		Timestamp:    0, // Será establecido por el servidor
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = kw.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("weather"),
		Value: msgBytes,
	})

	if err != nil {
		log.Printf("Error escribiendo a Kafka: %v", err)
	} else {
		log.Printf("Mensaje enviado a Kafka exitosamente")
	}

	return err
}

func (kw *KafkaWriter) Close() error {
	return kw.writer.Close()
}
