# Consumidor de Kafka para Valkey

Este componente consume mensajes del tópico de Kafka y almacena los datos en Valkey.

## Estructura del Proyecto

```
.
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── k8s/
│   └── deployment.yaml
└── README.md
```

## Funcionalidades

- Consume mensajes del tópico de Kafka
- Procesa los datos y los almacena en Valkey
- Implementa métricas con Prometheus
- Maneja la concurrencia con goroutines
- Implementa health checks para Kubernetes

## Compilación y Ejecución Local

```bash
go mod tidy
go build -o kafka-consumer .
./kafka-consumer
```

## Compilación y Subida al Registry

```bash
docker build -t kafka-consumer:latest .
docker tag kafka-consumer:latest <IP-ZOT>:5000/202200129/kafka-consumer:latest
docker push <IP-ZOT>:5000/202200129/kafka-consumer:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```