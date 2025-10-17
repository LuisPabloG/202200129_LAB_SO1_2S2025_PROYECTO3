# Manifiestos de Kubernetes para el Proyecto 3

Este directorio contiene los archivos YAML necesarios para desplegar todos los componentes del proyecto en Kubernetes.

## Archivos

1. `namespace.yaml`: Crea el namespace para el proyecto
2. `ingress.yaml`: Configura el Ingress para el tráfico entrante
3. `kafka-strimzi.yaml`: Configuración de Kafka usando Strimzi
4. `rabbitmq.yaml`: Configuración de RabbitMQ
5. `valkey.yaml`: Configuración de Valkey (base de datos en memoria)
6. `grafana.yaml`: Configuración de Grafana para visualización

## Despliegue

Para desplegar todos los componentes en el clúster de Kubernetes:

```bash
kubectl apply -f namespace.yaml
kubectl apply -f kafka-strimzi.yaml
kubectl apply -f rabbitmq.yaml
kubectl apply -f valkey.yaml
kubectl apply -f ingress.yaml
kubectl apply -f grafana.yaml
```