#!/bin/bash
# Script para acceder y hacer queries a Valkey
# Carnet: 202200129 | Municipio: Chinautla

set -e

echo "╔════════════════════════════════════════════════════════════════════════╗"
echo "║                    ACCESO A VALKEY (REDIS)                            ║"
echo "║                    Carnet: 202200129 - Chinautla                      ║"
echo "╚════════════════════════════════════════════════════════════════════════╝"
echo ""

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Menu
echo -e "${BLUE}════ OPCIONES ════${NC}"
echo ""
echo "1. Acceso interactivo a redis-cli"
echo "2. Ver todas las claves (KEYS weather:*)"
echo "3. Ver contadores de clima"
echo "4. Ver datos detallados por municipio"
echo "5. Ver estadísticas completas"
echo "6. Monitorear en tiempo real"
echo "7. Limpiar datos (FLUSH - PELIGRO)"
echo "8. Salir"
echo ""
read -p "Selecciona una opción (1-8): " option

case $option in
    1)
        echo ""
        echo -e "${BLUE}Conectando a Valkey...${NC}"
        echo "Escribe 'exit' para salir"
        echo ""
        kubectl exec -it valkey-0 -n weather-system -- redis-cli
        ;;
    2)
        echo ""
        echo -e "${MAGENTA}═══ TODAS LAS CLAVES ═══${NC}"
        echo ""
        kubectl exec -it valkey-0 -n weather-system -- redis-cli KEYS "weather:*"
        echo ""
        ;;
    3)
        echo ""
        echo -e "${MAGENTA}═══ CONTADORES DE CLIMA ═══${NC}"
        echo ""
        kubectl exec valkey-0 -n weather-system -- redis-cli << EOF
GET weather:sunny
GET weather:cloudy
GET weather:rainy
GET weather:foggy
EOF
        echo ""
        ;;
    4)
        echo ""
        echo -e "${MAGENTA}═══ DATOS DETALLADOS POR MUNICIPIO ═══${NC}"
        echo ""
        kubectl exec valkey-0 -n weather-system -- redis-cli << EOF
HGETALL weather:data:chinautla:sunny
HGETALL weather:data:chinautla:cloudy
HGETALL weather:data:chinautla:rainy
HGETALL weather:data:chinautla:foggy
EOF
        echo ""
        ;;
    5)
        echo ""
        echo -e "${MAGENTA}═══ ESTADÍSTICAS COMPLETAS ═══${NC}"
        echo ""
        kubectl exec valkey-0 -n weather-system -- redis-cli INFO
        echo ""
        ;;
    6)
        echo ""
        echo -e "${BLUE}Monitoreando Valkey...${NC}"
        echo "Presiona Ctrl+C para salir"
        echo ""
        timeout 30 kubectl exec valkey-0 -n weather-system -- redis-cli MONITOR || true
        echo ""
        ;;
    7)
        echo ""
        echo -e "${RED}⚠️  PELIGRO: Esto eliminará TODOS los datos${NC}"
        read -p "¿Estás seguro? (escribe 'SI' para confirmar): " confirm
        if [ "$confirm" = "SI" ]; then
            echo ""
            echo -e "${BLUE}Limpiando datos...${NC}"
            kubectl exec valkey-0 -n weather-system -- redis-cli FLUSHALL
            echo -e "${GREEN}✓ Datos eliminados${NC}"
        else
            echo "Operación cancelada"
        fi
        echo ""
        ;;
    8)
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
