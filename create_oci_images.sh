#!/bin/bash

# Script para crear y configurar las imágenes OCI para el Proyecto 3
# Estudiante: 202200129

set -e  # Exit on any error

# Obtener la IP del registro Zot
REGISTRY_IP=$(gcloud compute instances describe zot-registry-202200129 --zone=us-central1-a --project proyecto-3-475405 --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
echo "La IP del registro Zot es: $REGISTRY_IP"

echo "===== Generando imágenes OCI para Proyecto 3 - 202200129 ====="

# Verificar existencia de Docker
if ! command -v docker &> /dev/null; then
    echo "Docker no está instalado. Instalando Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    echo "Docker instalado. Por favor cierre sesión y vuelva a iniciar para usar Docker sin sudo."
    exit 1
fi

# Función para construir y etiquetar imágenes
build_image() {
    local service_name=$1
    local service_path=$2
    local version=$3
    
    echo "Construyendo imagen para $service_name..."
    
    # Construcción de la imagen
    docker build -t 202200129/$service_name:$version $service_path
    
    # Etiquetar para Zot Registry
    docker tag 202200129/$service_name:$version ${REGISTRY_IP}:5000/202200129/$service_name:$version
    
    echo "Imagen $service_name:$version construida y etiquetada correctamente."
}

# Definir versión
VERSION="1.0.0"

# Construir imágenes para todos los servicios
build_image "rust-api" "./2_rust_api" $VERSION
build_image "go-api" "./3_go_api" $VERSION
build_image "go-kafka-writer" "./4_go_kafka_writer" $VERSION
build_image "go-rabbit-writer" "./5_go_rabbit_writer" $VERSION
build_image "go-kafka-consumer" "./6_go_kafka_consumer" $VERSION
build_image "go-rabbit-consumer" "./7_go_rabbit_consumer" $VERSION

echo "===== Publicando imágenes en el registro Zot ====="

# Verificar que Zot esté ejecutando en la IP remota en el puerto 5000
if ! curl -s http://${REGISTRY_IP}:5000/v2/ > /dev/null; then
    echo "No se pudo conectar al registro Zot en ${REGISTRY_IP}:5000."
    echo "Por favor asegúrese de que el registro esté en ejecución en la VM de GCP."
    exit 1
fi

# Función para publicar imágenes en Zot
push_to_zot() {
    local service_name=$1
    local version=$2
    
    echo "Publicando imagen $service_name:$version en Zot..."
    docker push ${REGISTRY_IP}:5000/202200129/$service_name:$version
}

# Publicar todas las imágenes en Zot
push_to_zot "rust-api" $VERSION
push_to_zot "go-api" $VERSION
push_to_zot "go-kafka-writer" $VERSION
push_to_zot "go-rabbit-writer" $VERSION
push_to_zot "go-kafka-consumer" $VERSION
push_to_zot "go-rabbit-consumer" $VERSION

echo "===== Todas las imágenes OCI han sido creadas y publicadas en Zot ====="
echo "Para verificar, visite: http://${REGISTRY_IP}:5000/v2/_catalog"