FROM golang:1.21 as builder
WORKDIR /app
COPY . .

# Descargar dependencias y compilar
RUN go mod tidy
RUN go build -o rabbit-consumer .

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/rabbit-consumer /usr/local/bin/rabbit-consumer

# Variables de entorno
ENV RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
ENV RABBITMQ_QUEUE=weather-tweets
ENV VALKEY_ADDR=valkey:6379
ENV DATA_TTL_HOURS=24

# Metadatos
LABEL org.opencontainers.image.authors="Luis Pablo Garcia - 202200129"
LABEL org.opencontainers.image.description="Consumidor de RabbitMQ para el Proyecto 3 de Sistemas Operativos 1"

EXPOSE 9093
CMD ["rabbit-consumer"]