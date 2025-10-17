#!/bin/bash

# Script para actualizar las referencias de imágenes en los archivos de deployment
# Estudiante: 202200129

set -e  # Exit on any error

# Obtener la IP del registro Zot
REGISTRY_IP=$(gcloud compute instances describe zot-registry-202200129 --zone=us-central1-a --project proyecto-3-475405 --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
echo "La IP del registro Zot es: $REGISTRY_IP"

# Actualizar las referencias de imágenes en los archivos de deployment
echo "Actualizando referencias de imágenes en los archivos de deployment..."

# Actualizar Rust API
sed -i "s|image: .*rust-api.*|image: ${REGISTRY_IP}:5000/202200129/rust-api:1.0.0|g" 2_rust_api/k8s/deployment.yaml

# Actualizar Go API
sed -i "s|image: .*go-api.*|image: ${REGISTRY_IP}:5000/202200129/go-api:1.0.0|g" 3_go_api/k8s/deployment.yaml

# Actualizar Kafka Writer
sed -i "s|image: .*kafka-writer.*|image: ${REGISTRY_IP}:5000/202200129/go-kafka-writer:1.0.0|g" 4_go_kafka_writer/k8s/deployment.yaml

# Actualizar RabbitMQ Writer
sed -i "s|image: .*rabbit-writer.*|image: ${REGISTRY_IP}:5000/202200129/go-rabbit-writer:1.0.0|g" 5_go_rabbit_writer/k8s/deployment.yaml

# Actualizar Kafka Consumer
sed -i "s|image: .*kafka-consumer.*|image: ${REGISTRY_IP}:5000/202200129/go-kafka-consumer:1.0.0|g" 6_go_kafka_consumer/k8s/deployment.yaml

# Actualizar RabbitMQ Consumer
sed -i "s|image: .*rabbit-consumer.*|image: ${REGISTRY_IP}:5000/202200129/go-rabbit-consumer:1.0.0|g" 7_go_rabbit_consumer/k8s/deployment.yaml

echo "Referencias de imágenes actualizadas correctamente."