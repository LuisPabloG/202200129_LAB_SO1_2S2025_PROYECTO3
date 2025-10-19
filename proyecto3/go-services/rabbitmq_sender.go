package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	pb "weather-go-services/proto"
)

type RabbitMQSender struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQSender(addr string) *RabbitMQSender {
	conn, err := amqp.Dial(addr)
	if err != nil {
		log.Printf("Error conectando a RabbitMQ: %v", err)
		return &RabbitMQSender{}
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Error abriendo canal: %v", err)
		conn.Close()
		return &RabbitMQSender{}
	}

	// Declarar exchange
	err = ch.ExchangeDeclare("weather-exchange", "direct", true, false, false, false, nil)
	if err != nil {
		log.Printf("Error declarando exchange: %v", err)
	}

	// Declarar queue
	q, err := ch.QueueDeclare("weather-tweets", true, false, false, false, nil)
	if err != nil {
		log.Printf("Error declarando queue: %v", err)
	}

	// Bind queue a exchange
	err = ch.QueueBind(q.Name, "weather", "weather-exchange", false, nil)
	if err != nil {
		log.Printf("Error haciendo bind: %v", err)
	}

	return &RabbitMQSender{
		conn:    conn,
		channel: ch,
	}
}

func (rs *RabbitMQSender) Send(req *pb.WeatherTweetRequest) error {
	if rs.channel == nil {
		return nil // No hacer nada si no hay conexión
	}

	msg := WeatherMessage{
		Municipality: req.Municipality,
		Temperature:  req.Temperature,
		Humidity:     req.Humidity,
		Weather:      req.Weather,
		Timestamp:    0,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = rs.channel.Publish(
		"weather-exchange",
		"weather",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)

	if err != nil {
		log.Printf("Error publicando a RabbitMQ: %v", err)
	} else {
		log.Printf("Mensaje enviado a RabbitMQ exitosamente")
	}

	return err
}

func (rs *RabbitMQSender) Close() error {
	if rs.channel != nil {
		rs.channel.Close()
	}
	if rs.conn != nil {
		return rs.conn.Close()
	}
	return nil
}
