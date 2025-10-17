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
RUN go build -o go-api .

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/go-api /usr/local/bin/go-api

# Variables de entorno
ENV KAFKA_GRPC_SERVICE_ADDR=kafka-writer:50051
ENV RABBIT_GRPC_SERVICE_ADDR=rabbit-writer:50052

# Metadatos
LABEL org.opencontainers.image.authors="Luis Pablo Garcia - 202200129"
LABEL org.opencontainers.image.description="API REST y Cliente gRPC en Go para el Proyecto 3 de Sistemas Operativos 1"

EXPOSE 8080
CMD ["go-api"]