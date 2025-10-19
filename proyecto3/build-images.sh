#!/bin/bash
# Script para compilar imágenes Docker

set -e

ZOT_REGISTRY="${1:-34.159.50.100:5000}"

echo "========================================="
echo "Compilación de imágenes Docker"
echo "========================================="
echo "Registro: $ZOT_REGISTRY"
echo ""

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# API Rust
echo -e "${YELLOW}→ Compilando API Rust...${NC}"
cd api-rust
docker build -t $ZOT_REGISTRY/weather-api-rust:latest .
docker push $ZOT_REGISTRY/weather-api-rust:latest
echo -e "${GREEN}✓ API Rust compilada y enviada${NC}"
echo ""

# Go Deployment 1
echo -e "${YELLOW}→ Compilando Go Deployment 1...${NC}"
cd ../go-services
docker build -f Dockerfile.deployment1 -t $ZOT_REGISTRY/go-services-deployment1:latest .
docker push $ZOT_REGISTRY/go-services-deployment1:latest
echo -e "${GREEN}✓ Go Deployment 1 compilado y enviado${NC}"
echo ""

# Kafka Consumer
echo -e "${YELLOW}→ Compilando Kafka Consumer...${NC}"
docker build -f Dockerfile.kafka-consumer -t $ZOT_REGISTRY/go-kafka-consumer:latest .
docker push $ZOT_REGISTRY/go-kafka-consumer:latest
echo -e "${GREEN}✓ Kafka Consumer compilado y enviado${NC}"
echo ""

# RabbitMQ Consumer
echo -e "${YELLOW}→ Compilando RabbitMQ Consumer...${NC}"
docker build -f Dockerfile.rabbitmq-consumer -t $ZOT_REGISTRY/go-rabbitmq-consumer:latest .
docker push $ZOT_REGISTRY/go-rabbitmq-consumer:latest
echo -e "${GREEN}✓ RabbitMQ Consumer compilado y enviado${NC}"
echo ""

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}✓ Todas las imágenes compiladas y enviadas${NC}"
echo -e "${GREEN}=========================================${NC}"
