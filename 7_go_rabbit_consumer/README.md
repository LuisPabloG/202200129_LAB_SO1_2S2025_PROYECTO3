# Consumidor de RabbitMQ para Valkey

Este componente consume mensajes de la cola de RabbitMQ y almacena los datos en Valkey.

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

- Consume mensajes de la cola de RabbitMQ
- Procesa los datos y los almacena en Valkey
- Implementa métricas con Prometheus
- Maneja la concurrencia con goroutines
- Implementa health checks para Kubernetes

## Compilación y Ejecución Local

```bash
go mod tidy
go build -o rabbit-consumer .
./rabbit-consumer
```

## Compilación y Subida al Registry

```bash
docker build -t rabbit-consumer:latest .
docker tag rabbit-consumer:latest <IP-ZOT>:5000/202200129/rabbit-consumer:latest
docker push <IP-ZOT>:5000/202200129/rabbit-consumer:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```