#!/bin/bash
# Script para ejecutar todo el flujo de pruebas - Proyecto 3
# Carnet: 202200129 | Municipio: Chinautla

set -e

echo "╔════════════════════════════════════════════════════════════════════════╗"
echo "║             PRUEBAS DE CARGA - SISTEMA DE TWEETS DEL CLIMA            ║"
echo "║                    Carnet: 202200129 - Chinautla                      ║"
echo "╚════════════════════════════════════════════════════════════════════════╝"
echo ""

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# PASO 1: Verificar status
echo -e "${BLUE}════ PASO 1: Verificar Status del Cluster ════${NC}"
echo ""
echo "Ejecutando: kubectl get all -n weather-system"
echo ""
kubectl get all -n weather-system 2>/dev/null || echo "Namespace no existe aún"
echo ""

# PASO 2: Obtener IP del Ingress
echo -e "${BLUE}════ PASO 2: Obtener IP del Ingress ════${NC}"
INGRESS_IP=$(kubectl get ingress weather-ingress -n weather-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

if [ -z "$INGRESS_IP" ]; then
    echo -e "${YELLOW}⚠️  IP del Ingress aún no asignada${NC}"
else
    echo "IP del Ingress: $INGRESS_IP"
fi
echo ""

# PASO 3: Mostrar opciones de acceso
echo -e "${BLUE}════ PASO 3: Acceso a Servicios ════${NC}"
echo ""
echo -e "${GREEN}✓ LOCUST${NC} (Generador de Carga)"
echo "  Ejecuta: ./run-locust.sh"
echo ""
echo -e "${GREEN}✓ GRAFANA${NC} (Dashboard)"
echo "  Ejecuta: ./run-grafana.sh"
echo ""
echo -e "${GREEN}✓ VALKEY${NC} (Base de Datos)"
echo "  Ejecuta: ./run-valkey.sh"
echo ""

# PASO 4: Mostrar comandos útiles
echo -e "${BLUE}════ COMANDOS ÚTILES ════${NC}"
echo ""
echo "Ver status en vivo:"
echo "  kubectl get pods -n weather-system -w"
echo ""
echo "Ver HPA escalando:"
echo "  kubectl get hpa -n weather-system -w"
echo ""
echo "Ver logs de servicios:"
echo "  kubectl logs -f deployment/go-deployment-1 -n weather-system"
echo ""

echo -e "${GREEN}✓ Sistema listo para pruebas${NC}"
