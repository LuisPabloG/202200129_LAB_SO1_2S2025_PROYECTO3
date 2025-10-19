#!/bin/bash
# Script para configurar y acceder a Grafana
# Carnet: 202200129 | Municipio: Chinautla

set -e

echo "╔════════════════════════════════════════════════════════════════════════╗"
echo "║                     CONFIGURACIÓN DE GRAFANA                          ║"
echo "║                    Carnet: 202200129 - Chinautla                      ║"
echo "╚════════════════════════════════════════════════════════════════════════╝"
echo ""

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Obtener IP del Ingress
echo -e "${BLUE}Obteniendo acceso a Grafana...${NC}"
INGRESS_IP=$(kubectl get ingress weather-ingress -n weather-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

if [ -z "$INGRESS_IP" ]; then
    echo -e "${YELLOW}⚠️  IP del Ingress aún no disponible${NC}"
    echo "Intentando con port-forward..."
    echo ""
    echo -e "${BLUE}Abriendo puerto 3000...${NC}"
    kubectl port-forward svc/grafana 3000:3000 -n weather-system 2>/dev/null &
    PORT_FORWARD_PID=$!
    sleep 3
    INGRESS_IP="localhost:3000"
    trap "kill $PORT_FORWARD_PID 2>/dev/null" EXIT
else
    echo -e "${GREEN}✓ IP obtenida: $INGRESS_IP${NC}"
fi

echo ""
echo -e "${BLUE}════ CREDENCIALES ════${NC}"
echo ""
echo "Usuario: admin"
echo "Contraseña: admin123"
echo ""

# Menu de opciones
echo -e "${BLUE}════ OPCIONES ════${NC}"
echo ""
echo "1. Abrir Grafana en navegador"
echo "2. Ver URL de acceso"
echo "3. Ver status del pod"
echo "4. Ver logs de Grafana"
echo "5. Salir"
echo ""
read -p "Selecciona una opción (1-5): " option

case $option in
    1)
        echo ""
        if command -v xdg-open &> /dev/null; then
            xdg-open "http://$INGRESS_IP" 2>/dev/null &
        elif command -v open &> /dev/null; then
            open "http://$INGRESS_IP" 2>/dev/null &
        else
            echo -e "${YELLOW}Abre en tu navegador: http://$INGRESS_IP${NC}"
        fi
        echo -e "${GREEN}✓ Abierto en navegador${NC}"
        echo "Esperando 5 segundos..."
        sleep 5
        ;;
    2)
        echo ""
        echo -e "${GREEN}✓ URL de Grafana:${NC}"
        echo "  http://$INGRESS_IP"
        echo ""
        ;;
    3)
        echo ""
        echo "Status del pod Grafana:"
        kubectl get pod -n weather-system -l app=grafana
        echo ""
        ;;
    4)
        echo ""
        echo "Logs de Grafana (últimas 20 líneas):"
        echo ""
        kubectl logs deployment/grafana -n weather-system --tail=20
        echo ""
        ;;
    5)
        echo "¡Hasta luego!"
        exit 0
        ;;
    *)
        echo -e "${RED}Opción inválida${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✓ Proceso completado${NC}"
