#!/bin/bash
# Script para ejecutar Locust contra el cluster
# Carnet: 202200129 | Municipio: Chinautla

set -e

echo "╔════════════════════════════════════════════════════════════════════════╗"
echo "║                    GENERADOR DE CARGA CON LOCUST                      ║"
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
echo -e "${BLUE}Obteniendo IP del Ingress...${NC}"
INGRESS_IP=$(kubectl get ingress weather-ingress -n weather-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

if [ -z "$INGRESS_IP" ]; then
    echo -e "${RED}❌ Error: No se pudo obtener la IP del Ingress${NC}"
    echo "Verifica que el Ingress esté creado:"
    echo "  kubectl get ingress -n weather-system"
    exit 1
fi

echo -e "${GREEN}✓ IP obtenida: $INGRESS_IP${NC}"
echo ""

# Verificar si Locust está instalado
echo -e "${BLUE}Verificando Locust...${NC}"
if ! command -v locust &> /dev/null; then
    echo -e "${YELLOW}⚠️  Locust no está instalado. Instalando...${NC}"
    pip3 install locust requests 2>&1 | tail -5
fi
echo -e "${GREEN}✓ Locust listo${NC}"
echo ""

# Menu de opciones
echo -e "${BLUE}════ OPCIONES DE CARGA ════${NC}"
echo ""
echo "1. Modo Web (Interface Interactiva) - RECOMENDADO"
echo "2. Modo Headless Ligero (1000 tweets, 5 usuarios)"
echo "3. Modo Headless Medio (5000 tweets, 10 usuarios)"
echo "4. Modo Headless Pesado (10000 tweets, 20 usuarios)"
echo "5. Personalizado (especificar número de usuarios y tweets)"
echo ""
read -p "Selecciona una opción (1-5): " option

cd /home/luis-pablo-garcia/Escritorio/PROYECTO\ 3\ SOPES/202200129_LAB_SO1_2S2025_PROYECTO3/proyecto3/locust-config

case $option in
    1)
        echo ""
        echo -e "${BLUE}Iniciando Locust en modo WEB...${NC}"
        echo -e "${YELLOW}Interface disponible en: http://localhost:8089${NC}"
        echo "Presiona Ctrl+C para detener"
        echo ""
        locust -f locustfile.py --host=http://$INGRESS_IP --users 10 --spawn-rate 5
        ;;
    2)
        echo ""
        echo -e "${BLUE}Iniciando con 1000 tweets y 5 usuarios...${NC}"
        locust -f locustfile.py \
            --host=http://$INGRESS_IP \
            --headless \
            -u 5 \
            -r 5 \
            -n 1000
        ;;
    3)
        echo ""
        echo -e "${BLUE}Iniciando con 5000 tweets y 10 usuarios...${NC}"
        locust -f locustfile.py \
            --host=http://$INGRESS_IP \
            --headless \
            -u 10 \
            -r 5 \
            -n 5000
        ;;
    4)
        echo ""
        echo -e "${BLUE}Iniciando con 10000 tweets y 20 usuarios...${NC}"
        locust -f locustfile.py \
            --host=http://$INGRESS_IP \
            --headless \
            -u 20 \
            -r 10 \
            -n 10000
        ;;
    5)
        read -p "¿Cuántos usuarios concurrentes? " users
        read -p "¿Cuántos tweets a generar? " tweets
        echo ""
        echo -e "${BLUE}Iniciando con $tweets tweets y $users usuarios...${NC}"
        locust -f locustfile.py \
            --host=http://$INGRESS_IP \
            --headless \
            -u $users \
            -r $users \
            -n $tweets
        ;;
    *)
        echo -e "${RED}Opción inválida${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✓ Ejecución completada${NC}"
