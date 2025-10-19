# Proyecto 3 - Sistema de Tweets del Clima

Arquitectura distribuida en Kubernetes (GKE) para procesar tweets de clima en tiempo real.

**Carnet:** 202200129  
**Estudiante:** Luis Pablo Manuel García López  
**Municipio:** Chinautla

## Estructura del Proyecto

```
proyecto3/
├── proto/                   # Archivos de Protocol Buffers
│   └── weather_tweet.proto
├── api-rust/               # API REST en Rust
│   ├── src/
│   ├── Cargo.toml
│   ├── build.rs
│   └── Dockerfile
├── go-services/            # Servicios en Go
│   ├── deployment1.go      # Server gRPC + Clients
│   ├── kafka_writer.go     # Writer para Kafka
│   ├── kafka_consumer.go   # Consumer de Kafka
│   ├── rabbitmq_sender.go  # Sender para RabbitMQ
│   ├── rabbitmq_consumer.go# Consumer de RabbitMQ
│   ├── go.mod
│   └── Dockerfile.*
├── k8s-manifests/          # YAMLs de Kubernetes
│   ├── 01-namespace.yaml
│   ├── 02-rust-api.yaml
│   ├── 03-go-deployment-1.yaml
│   ├── 04-consumers.yaml
│   ├── 05-valkey.yaml
│   ├── 06-kafka.yaml
│   ├── 07-rabbitmq.yaml
│   ├── 08-ingress.yaml
│   └── 09-grafana.yaml
├── locust-config/          # Configuración de Locust
│   ├── locustfile.py
│   └── requirements.txt
├── TECNICA.md              # Documentación técnica completa
└── README.md               # Este archivo
```

## Requisitos

- Docker
- Kubernetes 1.24+
- kubectl configurado
- GCP con GKE
- Rust 1.70+ (para compilar API)
- Go 1.21+ (para compilar servicios)
- Python 3.9+ (para Locust)

## Inicio Rápido

### 1. Configurar credenciales GCP

```bash
gcloud container clusters get-credentials proyecto3-sopes-1 \
  --zone us-central1-c \
  --project proyecto-3-475405
```

### 2. Crear namespace

```bash
kubectl apply -f proyecto3/k8s-manifests/01-namespace.yaml
```

### 3. Desplegar infraestructura base

```bash
# Valkey, Kafka, RabbitMQ
kubectl apply -f proyecto3/k8s-manifests/05-valkey.yaml
kubectl apply -f proyecto3/k8s-manifests/06-kafka.yaml
kubectl apply -f proyecto3/k8s-manifests/07-rabbitmq.yaml

# Esperar 5-10 minutos
kubectl get pods -n weather-system
```

### 4. Construir y enviar imágenes

```bash
# Reemplazar con IP de Zot Registry
export ZOT_REGISTRY="34.159.50.100:5000"

# API Rust
cd proyecto3/api-rust
docker build -t $ZOT_REGISTRY/weather-api-rust:latest .
docker push $ZOT_REGISTRY/weather-api-rust:latest

# Go Services
cd ../go-services
docker build -f Dockerfile.deployment1 -t $ZOT_REGISTRY/go-services-deployment1:latest .
docker push $ZOT_REGISTRY/go-services-deployment1:latest
# ... (repetir para kafka y rabbitmq consumers)
```

### 5. Actualizar YAMLs

Reemplazar `ZOT_REGISTRY` en los archivos YAML con tu IP.

### 6. Desplegar servicios

```bash
kubectl apply -f proyecto3/k8s-manifests/02-rust-api.yaml
kubectl apply -f proyecto3/k8s-manifests/03-go-deployment-1.yaml
kubectl apply -f proyecto3/k8s-manifests/04-consumers.yaml
kubectl apply -f proyecto3/k8s-manifests/09-grafana.yaml
kubectl apply -f proyecto3/k8s-manifests/08-ingress.yaml
```

### 7. Verificar estado

```bash
kubectl get all -n weather-system
kubectl get ingress -n weather-system
```

## Pruebas de Carga

```bash
cd proyecto3/locust-config
pip install -r requirements.txt

# Modo UI (interactivo)
locust -f locustfile.py --host=http://INGRESS_IP

# Modo headless (10K requests, 10 usuarios)
locust -f locustfile.py \
  --host=http://INGRESS_IP \
  --headless \
  -u 10 \
  -r 10 \
  -n 10000
```

## Acceso a Servicios

| Servicio | URL/Acceso | Notas |
|----------|-----------|-------|
| API REST | `http://weather.local/api/tweets` | Ingress |
| Grafana | `http://grafana.local` | Usuario: admin / Pass: admin123 |
| RabbitMQ | `kubectl port-forward svc/rabbitmq-management 15672:15672` | localhost:15672 |
| Valkey | `kubectl exec -it valkey-0 -- redis-cli` | CLI interactiva |

## Monitoreo

```bash
# Ver logs de Deployment 1
kubectl logs -f deployment/go-deployment-1 -n weather-system

# Ver HPA en acción
kubectl get hpa -n weather-system -w

# Ver eventos
kubectl get events -n weather-system --sort-by='.lastTimestamp'
```

## Documentación Completa

Ver `TECNICA.md` para:
- Arquitectura detallada
- Análisis de rendimiento
- Respuestas a preguntas técnicas
- Instrucciones de configuración avanzada

## Notas Importantes

1. **Para la calificación:**
   - Valkey debe estar **vacío** al inicio
   - Clúster debe estar corriendo 24 horas antes
   - Todas las imágenes deben estar en Zot Registry

2. **Municipio asignado:** Chinautla (carnet termina en 9)

3. **Dashboard:** Gráfica de barras con conteos por clima

## Comandos Útiles

```bash
# Limpiar todo
kubectl delete namespace weather-system

# Recrear desde cero
kubectl apply -f proyecto3/k8s-manifests/01-namespace.yaml
# ... (aplicar todos los YAML en orden)

# Ver recursos
kubectl top nodes
kubectl top pods -n weather-system

# Debug de un pod
kubectl exec -it pod/NAME -n weather-system -- sh
```

## Contacto

Para preguntas o issues: Luis Pablo García López (202200129)

---

**Última actualización:** Octubre 18, 2025
