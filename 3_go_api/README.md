# Servicio Go - API REST y Cliente gRPC

Este componente es una API REST en Go que recibe peticiones de la API Rust y funciona como cliente gRPC para enviar los datos a los servicios de publicación en Kafka y RabbitMQ.

## Estructura del Proyecto

```
.
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── proto/
│   └── weathertweet.proto
├── k8s/
│   └── deployment.yaml
└── README.md
```

## Funcionalidades

- Recibe peticiones HTTP de la API Rust
- Actúa como cliente gRPC para enviar mensajes a los servicios de publicación
- Maneja la concurrencia con goroutines
- Implementa health checks para Kubernetes

## Compilación y Ejecución Local

```bash
go mod tidy
go build -o go-api .
./go-api
```

## Compilación y Subida al Registry

```bash
docker build -t go-api:latest .
docker tag go-api:latest <IP-ZOT>:5000/202200129/go-api:latest
docker push <IP-ZOT>:5000/202200129/go-api:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```