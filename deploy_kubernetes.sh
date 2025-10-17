#!/bin/bash

# Script para desplegar el Proyecto 3 en Kubernetes (GKE)
# Estudiante: 202200129

set -e  # Exit on any error

echo "===== Desplegando Proyecto 3 - 202200129 en Kubernetes ====="

# Verificar que kubectl esté instalado
if ! command -v kubectl &> /dev/null; then
    echo "kubectl no está instalado. Por favor instálelo para continuar."
    exit 1
fi

# Verificar que gcloud esté instalado
if ! command -v gcloud &> /dev/null; then
    echo "gcloud no está instalado. Por favor instálelo para continuar."
    exit 1
fi

# Verificar que estamos conectados al cluster correcto
CURRENT_CLUSTER=$(kubectl config current-context)
EXPECTED_CLUSTER="gke-sopes3-202200129"

if [[ "$CURRENT_CLUSTER" != *"$EXPECTED_CLUSTER"* ]]; then
    echo "Conectando al cluster GKE gke-sopes3-202200129..."
    gcloud container clusters get-credentials gke-sopes3-202200129 --zone us-central1-a --project proyecto-3-475405
fi

echo "===== Aplicando configuraciones de Kubernetes ====="

# Crear namespace
kubectl apply -f k8s/namespace.yaml

# Desplegar la infraestructura
echo "Desplegando infraestructura..."
kubectl apply -f k8s/kafka-strimzi.yaml
kubectl apply -f k8s/rabbitmq.yaml
kubectl apply -f k8s/valkey.yaml
kubectl apply -f k8s/prometheus.yaml
kubectl apply -f k8s/grafana-config.yaml
kubectl apply -f k8s/grafana.yaml

# Esperar a que la infraestructura esté lista
echo "Esperando que los componentes de infraestructura estén listos..."
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=kafka --timeout=300s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=rabbitmq --timeout=300s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=valkey --timeout=300s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=prometheus --timeout=300s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=grafana --timeout=300s

# Desplegar los microservicios en orden
echo "Desplegando microservicios..."

# Primero los consumidores para que estén listos para recibir mensajes
kubectl apply -f 6_go_kafka_consumer/k8s/deployment.yaml
kubectl apply -f 7_go_rabbit_consumer/k8s/deployment.yaml

# Después los escritores de mensajes
kubectl apply -f 4_go_kafka_writer/k8s/deployment.yaml
kubectl apply -f 5_go_rabbit_writer/k8s/deployment.yaml

# Finalmente la API Go y Rust
kubectl apply -f 3_go_api/k8s/deployment.yaml
kubectl apply -f 2_rust_api/k8s/deployment.yaml

# Desplegar Locust para pruebas de carga
echo "Desplegando Locust para pruebas de carga..."
kubectl apply -f k8s/locust.yaml

# Aplicar configuración de Ingress
echo "Configurando Ingress para acceso externo..."
kubectl apply -f k8s/ingress.yaml

# Esperar a que todos los pods estén listos
echo "Esperando que todos los servicios estén listos..."
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=rust-api --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=go-api --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=kafka-writer --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=rabbit-writer --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=kafka-consumer --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=rabbit-consumer --timeout=180s
kubectl wait --namespace=proyecto3-202200129 --for=condition=Ready pod -l app=locust --timeout=180s

# Obtener la IP externa del Ingress
echo "Obteniendo dirección IP del Ingress..."
INGRESS_IP=""
while [ -z "$INGRESS_IP" ]; do
    INGRESS_IP=$(kubectl -n proyecto3-202200129 get ingress -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}')
    if [ -z "$INGRESS_IP" ]; then
        echo "Esperando asignación de IP para Ingress..."
        sleep 10
    fi
done

echo "===== Despliegue completado exitosamente ====="
echo "Sistema desplegado en GKE: gke-sopes3-202200129"
echo ""
echo "Acceso a servicios:"
echo "- Rust API: http://${INGRESS_IP}/tweet"
echo "- Grafana Dashboard: http://${INGRESS_IP}/grafana (usuario: admin, contraseña: admin)"
echo "- Locust (Pruebas de carga): http://${INGRESS_IP}/locust"
echo ""
echo "Carnet: 202200129"
echo "Municipio asignado: chinautla"