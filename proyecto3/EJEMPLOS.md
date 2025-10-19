### Ejemplos de Uso del Sistema

## 1. Request Manual con curl

```bash
# Obtener IP del Ingress
INGRESS_IP=$(kubectl get ingress weather-ingress -n weather-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Enviar un tweet
curl -X POST http://$INGRESS_IP/api/tweets \
  -H "Content-Type: application/json" \
  -d '{
    "municipality": "chinautla",
    "temperature": 25,
    "humidity": 65,
    "weather": "sunny"
  }'
```

## 2. Ejecutar Locust

```bash
cd locust-config

# Instalación de dependencias
pip install -r requirements.txt

# Interfaz web (abre http://localhost:8089)
locust -f locustfile.py --host=http://$INGRESS_IP

# Modo headless (10K requests)
locust -f locustfile.py \
  --host=http://$INGRESS_IP \
  --headless \
  -u 10 \
  -r 10 \
  -n 10000
```

## 3. Acceder a Grafana

```bash
# Obtener IP de Ingress
INGRESS_IP=$(kubectl get ingress weather-ingress -n weather-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Agregar a /etc/hosts (en Linux/Mac)
echo "$INGRESS_IP grafana.local" | sudo tee -a /etc/hosts

# Acceder a Grafana
# URL: http://grafana.local
# Usuario: admin
# Contraseña: admin123
```

## 4. Consultar Valkey

```bash
# Acceso interactivo a Valkey
kubectl exec -it valkey-0 -n weather-system -- redis-cli

# Dentro de redis-cli:
> KEYS weather:*
> GET weather:sunny
> GET weather:cloudy
> GET weather:rainy
> GET weather:foggy
> HGETALL weather:data:chinautla:sunny
```

## 5. Ver Logs en Tiempo Real

```bash
# Go Deployment 1
kubectl logs -f deployment/go-deployment-1 -n weather-system

# Kafka Consumer
kubectl logs -f deployment/kafka-consumer -n weather-system

# RabbitMQ Consumer
kubectl logs -f deployment/rabbitmq-consumer -n weather-system

# API Rust
kubectl logs -f deployment/rust-api -n weather-system
```

## 6. Escalar Manualmente

```bash
# Escalar API Rust
kubectl scale deployment rust-api -n weather-system --replicas=3

# Ver HPA en acción
kubectl get hpa -n weather-system -w
```

## 7. Acceso a RabbitMQ Management

```bash
# Port forwarding
kubectl port-forward svc/rabbitmq-management -n weather-system 15672:15672

# Acceder en navegador
# URL: http://localhost:15672
# Usuario: guest
# Contraseña: guest
```

## 8. Crear OCI Artifact en Zot

```bash
# Crear archivo de entrada (ejemplo)
echo '{"description": "Weather Tweet System Configuration"}' > config.json

# Subir como OCI Artifact
curl -X POST \
  -H "Content-Type: application/json" \
  -d @config.json \
  http://ZOT_IP:5000/v2/weather-config/blobs/uploads/

# En Dockerfile/Pod, descargar desde Zot
# Ver TECNICA.md para más detalles
```

## 9. Verificar Kafka Topics

```bash
# Dentro de Kafka pod
kubectl exec -it kafka-0 -n weather-system -- /bin/bash

# Listar topics
kafka-topics --list --bootstrap-server kafka:9092

# Ver mensajes en topic
kafka-console-consumer --bootstrap-server kafka:9092 \
  --topic weather-tweets \
  --from-beginning \
  --max-messages 10
```

## 10. Monitoreo del Sistema

```bash
# Script de monitoreo en tiempo real
./monitor.sh

# Ver eventos
kubectl get events -n weather-system --sort-by='.lastTimestamp'

# Ver métricas de nodos
kubectl top nodes

# Ver métricas de pods
kubectl top pods -n weather-system
```

---

## Troubleshooting

### Los pods no inician
```bash
kubectl describe pod POD_NAME -n weather-system
kubectl logs POD_NAME -n weather-system
```

### No hay datos en Valkey
```bash
# Verificar que los consumidores están corriendo
kubectl get pods -n weather-system | grep consumer

# Ver logs de consumidor
kubectl logs deployment/kafka-consumer -n weather-system
```

### API no responde
```bash
# Verificar Rust API pod
kubectl exec -it deployment/rust-api -n weather-system -- sh

# Probar conexión a Go service
curl http://go-deployment-1:8081/health
```

### Grafico no muestra datos
```bash
# 1. Verificar que Grafana puede conectarse a Valkey
# 2. Crear data source: redis://valkey:6379
# 3. Importar dashboard o crear queries
```

