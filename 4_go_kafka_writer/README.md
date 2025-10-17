# Servicio Go - gRPC Server y Kafka Writer

Este componente es un servidor gRPC en Go que recibe peticiones del cliente gRPC y publica mensajes en Kafka.

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

- Implementa un servidor gRPC para recibir mensajes de la API Go
- Publica mensajes en un topic de Kafka
- Maneja la concurrencia con goroutines
- Implementa health checks para Kubernetes

## Compilación y Ejecución Local

```bash
go mod tidy
go build -o kafka-writer .
./kafka-writer
```

## Compilación y Subida al Registry

```bash
docker build -t kafka-writer:latest .
docker tag kafka-writer:latest <IP-ZOT>:5000/202200129/kafka-writer:latest
docker push <IP-ZOT>:5000/202200129/kafka-writer:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```