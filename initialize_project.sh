#!/bin/bash

# Script principal para desplegar el Proyecto 3 en GCP
# Estudiante: 202200129

set -e  # Exit on any error

echo "===== INICIALIZACIÓN DEL PROYECTO 3 - 202200129 EN GCP ====="

# Verificar herramientas necesarias
echo "Verificando herramientas necesarias..."
for cmd in gcloud docker kubectl sed; do
    if ! command -v $cmd &> /dev/null; then
        echo "ERROR: $cmd no está instalado. Por favor instálelo para continuar."
        exit 1
    fi
done

# Configurar proyecto GCP
echo "Configurando proyecto GCP..."
gcloud config set project proyecto-3-475405

# Paso 1: Configurar Zot Registry si no existe
echo "Verificando si la VM de Zot Registry existe..."
if ! gcloud compute instances describe zot-registry-202200129 --zone=us-central1-a &> /dev/null; then
    echo "Creando VM para Zot Registry..."
    gcloud compute instances create zot-registry-202200129 \
      --zone=us-central1-a \
      --machine-type=e2-small \
      --image-family=ubuntu-2004-lts \
      --image-project=ubuntu-os-cloud \
      --tags=http-server,https-server

    echo "Configurando firewall para permitir tráfico HTTP..."
    gcloud compute firewall-rules create allow-zot-registry \
      --direction=INGRESS \
      --action=ALLOW \
      --rules=tcp:5000 \
      --target-tags=http-server

    echo "Instalando Docker y Zot en la VM..."
    gcloud compute ssh zot-registry-202200129 --zone=us-central1-a -- "sudo apt update && sudo apt install -y docker.io && sudo systemctl enable docker && sudo systemctl start docker && sudo docker run -d -p 5000:5000 --restart always --name zot ghcr.io/project-zot/zot:latest"
    
    echo "Esperando a que Zot esté disponible..."
    sleep 10
fi

# Verificar que Zot está funcionando
REGISTRY_IP=$(gcloud compute instances describe zot-registry-202200129 --zone=us-central1-a --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
echo "La IP del registro Zot es: $REGISTRY_IP"

if ! curl -s http://${REGISTRY_IP}:5000/v2/ &> /dev/null; then
    echo "ERROR: No se pudo conectar al registro Zot. Verificando estado de la VM..."
    gcloud compute instances describe zot-registry-202200129 --zone=us-central1-a
    
    echo "Reiniciando el contenedor Zot..."
    gcloud compute ssh zot-registry-202200129 --zone=us-central1-a -- "sudo docker restart zot || sudo docker run -d -p 5000:5000 --restart always --name zot ghcr.io/project-zot/zot:latest"
    
    echo "Esperando a que Zot esté disponible..."
    sleep 10
    
    if ! curl -s http://${REGISTRY_IP}:5000/v2/ &> /dev/null; then
        echo "ERROR: No se pudo conectar al registro Zot después de reiniciarlo. Por favor verifique manualmente."
        exit 1
    fi
fi

echo "Zot Registry está funcionando correctamente en $REGISTRY_IP:5000"

# Paso 2: Verificar y conectar al cluster GKE
echo "Conectando al cluster GKE..."
gcloud container clusters get-credentials gke-sopes3-202200129 --zone=us-central1-a --project proyecto-3-475405

# Paso 3: Construir y publicar imágenes
echo "¿Desea construir y publicar las imágenes en el registro Zot? (s/n)"
read BUILD_IMAGES

if [[ "$BUILD_IMAGES" == "s" ]]; then
    echo "Ejecutando script para construir y publicar imágenes..."
    ./create_oci_images.sh
else
    echo "Omitiendo construcción de imágenes."
fi

# Paso 4: Actualizar referencias de imágenes en archivos de despliegue
echo "Actualizando referencias de imágenes en archivos de despliegue..."
./update_image_references.sh

# Paso 5: Desplegar en Kubernetes
echo "Desplegando proyecto en Kubernetes..."
./deploy_kubernetes.sh

echo "===== PROYECTO 3 INICIALIZADO EXITOSAMENTE ====="
echo "Carnet: 202200129"
echo "Municipio asignado: chinautla"