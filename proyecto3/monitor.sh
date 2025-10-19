#!/bin/bash
# Script para monitorear el sistema

echo "========================================="
echo "Monitoreo del Sistema"
echo "========================================="
echo ""

# Colores
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}Estado de los Pods:${NC}"
kubectl get pods -n weather-system -o wide
echo ""

echo -e "${CYAN}Horizontal Pod Autoscaler:${NC}"
kubectl get hpa -n weather-system
echo ""

echo -e "${CYAN}Eventos recientes:${NC}"
kubectl get events -n weather-system --sort-by='.lastTimestamp' | tail -10
echo ""

echo -e "${CYAN}Uso de recursos:${NC}"
kubectl top pods -n weather-system 2>/dev/null || echo "(Métricas no disponibles)"
echo ""

echo "Presionar Ctrl+C para salir"
echo ""

# Monitoreo en tiempo real
watch -n 2 'kubectl get pods -n weather-system -o wide'
