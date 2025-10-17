FROM golang:1.21 as builder
WORKDIR /app
COPY . .

# Instalar el compilador de protocol buffers
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generar código Go a partir del archivo proto
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/weathertweet.proto

# Descargar dependencias y compilar
RUN go mod tidy
RUN go build -o kafka-writer .

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/kafka-writer /usr/local/bin/kafka-writer

# Variables de entorno
ENV KAFKA_BROKERS=kafka:9092
ENV KAFKA_TOPIC=weather-tweets

# Metadatos
LABEL org.opencontainers.image.authors="Luis Pablo Garcia - 202200129"
LABEL org.opencontainers.image.description="Servidor gRPC y Writer de Kafka en Go para el Proyecto 3 de Sistemas Operativos 1"

EXPOSE 50051
EXPOSE 9090
CMD ["kafka-writer"]