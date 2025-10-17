# Servicio Go - gRPC Server y RabbitMQ Writer

Este componente es un servidor gRPC en Go que recibe peticiones del cliente gRPC y publica mensajes en RabbitMQ.

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
- Publica mensajes en una cola de RabbitMQ
- Maneja la concurrencia con goroutines
- Implementa health checks para Kubernetes

## Compilación y Ejecución Local

```bash
go mod tidy
go build -o rabbit-writer .
./rabbit-writer
```

## Compilación y Subida al Registry

```bash
docker build -t rabbit-writer:latest .
docker tag rabbit-writer:latest <IP-ZOT>:5000/202200129/rabbit-writer:latest
docker push <IP-ZOT>:5000/202200129/rabbit-writer:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```