#!/bin/bash
# Script para actualizar YAMLs con dirección del Zot Registry

ZOT_REGISTRY="${1:-34.159.50.100:5000}"

echo "Actualizando YAMLs con Zot Registry: $ZOT_REGISTRY"

# Actualizar cada YAML
sed -i "s|ZOT_REGISTRY|$ZOT_REGISTRY|g" ./k8s-manifests/02-rust-api.yaml
sed -i "s|ZOT_REGISTRY|$ZOT_REGISTRY|g" ./k8s-manifests/03-go-deployment-1.yaml
sed -i "s|ZOT_REGISTRY|$ZOT_REGISTRY|g" ./k8s-manifests/04-consumers.yaml

echo "✓ YAMLs actualizados correctamente"
echo ""
echo "Para desplegar, ejecutar:"
echo "  ./deploy.sh"
