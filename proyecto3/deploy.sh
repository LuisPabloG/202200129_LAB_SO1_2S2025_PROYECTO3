#!/bin/bash
# Script para desplegar todo el sistema

set -e

echo "========================================="
echo "Despliegue del Sistema de Tweets - Clima"
echo "========================================="
echo ""

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Verificar que kubectl está disponible
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl no está instalado${NC}"
    exit 1
fi

# Verificar conexión a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Error: No hay conexión al cluster${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Conexión al cluster verificada${NC}"
echo ""

# Crear namespace
echo -e "${YELLOW}→ Creando namespace weather-system...${NC}"
kubectl apply -f ./k8s-manifests/01-namespace.yaml
echo -e "${GREEN}✓ Namespace creado${NC}"
echo ""

# Desplegar infraestructura base
echo -e "${YELLOW}→ Desplegando Valkey (base de datos)...${NC}"
kubectl apply -f ./k8s-manifests/05-valkey.yaml
echo -e "${GREEN}✓ Valkey desplegado${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando Kafka...${NC}"
kubectl apply -f ./k8s-manifests/06-kafka.yaml
echo -e "${GREEN}✓ Kafka desplegado${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando RabbitMQ...${NC}"
kubectl apply -f ./k8s-manifests/07-rabbitmq.yaml
echo -e "${GREEN}✓ RabbitMQ desplegado${NC}"
echo ""

# Esperar a que los recursos estén listos
echo -e "${YELLOW}→ Esperando a que los servicios estén listos (esto toma 5-10 minutos)...${NC}"
echo "  Presionar Ctrl+C para saltar"
kubectl wait --for=condition=ready pod -l app=valkey -n weather-system --timeout=600s 2>/dev/null || true
kubectl wait --for=condition=ready pod -l app=kafka -n weather-system --timeout=600s 2>/dev/null || true
kubectl wait --for=condition=ready pod -l app=rabbitmq -n weather-system --timeout=600s 2>/dev/null || true

echo ""
echo -e "${GREEN}✓ Servicios base listos${NC}"
echo ""

# Desplegar servicios principales
echo -e "${YELLOW}→ Desplegando API Rust...${NC}"
kubectl apply -f ./k8s-manifests/02-rust-api.yaml
echo -e "${GREEN}✓ API Rust desplegada${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando Go Deployment 1...${NC}"
kubectl apply -f ./k8s-manifests/03-go-deployment-1.yaml
echo -e "${GREEN}✓ Go Deployment 1 desplegado${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando Consumidores...${NC}"
kubectl apply -f ./k8s-manifests/04-consumers.yaml
echo -e "${GREEN}✓ Consumidores desplegados${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando Grafana...${NC}"
kubectl apply -f ./k8s-manifests/09-grafana.yaml
echo -e "${GREEN}✓ Grafana desplegado${NC}"
echo ""

echo -e "${YELLOW}→ Desplegando Ingress...${NC}"
kubectl apply -f ./k8s-manifests/08-ingress.yaml
echo -e "${GREEN}✓ Ingress desplegado${NC}"
echo ""

# Mostrar información final
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}✓ Despliegue completado exitosamente${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Comandos útiles:"
echo ""
echo "Ver estado de pods:"
echo "  kubectl get pods -n weather-system"
echo ""
echo "Ver logs:"
echo "  kubectl logs -f deployment/go-deployment-1 -n weather-system"
echo ""
echo "Ver HPA:"
echo "  kubectl get hpa -n weather-system -w"
echo ""
echo "Ingress IP:"
kubectl get ingress -n weather-system
echo ""
