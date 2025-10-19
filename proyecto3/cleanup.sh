#!/bin/bash
# Script para limpiar completamente el sistema

set -e

echo "========================================="
echo "Limpieza del Sistema"
echo "========================================="
echo ""

# Colores
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

read -p "¿Está seguro de que desea eliminar el namespace weather-system? (s/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Ss]$ ]]; then
    echo -e "${YELLOW}→ Eliminando namespace...${NC}"
    kubectl delete namespace weather-system
    echo -e "${GREEN}✓ Namespace eliminado${NC}"
    echo ""
    echo "Esperando a que se complete..."
    sleep 10
    echo -e "${GREEN}✓ Limpieza completada${NC}"
else
    echo "Operación cancelada"
fi
