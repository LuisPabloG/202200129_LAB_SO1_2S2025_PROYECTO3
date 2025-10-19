# Documentación Técnica - Proyecto 3: Sistema de Tweets del Clima

**Carnet:** 202200129  
**Estudiante:** Luis Pablo Manuel García López  
**Municipio Asignado:** Chinautla (último dígito 9)  
**Fecha:** Octubre 2025

---

## 1. Introducción

Este documento describe la arquitectura y la implementación del **Sistema de Tweets del Clima**, un sistema distribuido desplegado en Google Kubernetes Engine (GKE) que procesa datos de clima en tiempo real utilizando tecnologías modernas de microservicios.

---

## 2. Arquitectura General del Sistema

### 2.1 Componentes Principales

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENTE (Locust)                            │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP Requests (10,000 tweets)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Ingress Controller (NGINX)                     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│         API REST Rust (Puerto 8080)                             │
│         - Recibe tweets del clima                               │
│         - Valida estructura JSON                                │
│         - HPA: 1-3 réplicas según CPU > 30%                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │ gRPC
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│      Go Deployment 1 (Puerto 50051)                             │
│      - Servidor gRPC                                            │
│      - Publica en Kafka y RabbitMQ simultáneamente             │
└──────┬───────────────────────────────────────┬──────────────────┘
       │                                       │
       ▼ (Kafka Topic)                         ▼ (RabbitMQ Queue)
┌────────────────┐                      ┌────────────────┐
│     Kafka      │                      │    RabbitMQ    │
│   Broker       │                      │    Broker      │
└────────┬───────┘                      └────────┬───────┘
         │                                       │
         ▼ (Consumer)                            ▼ (Consumer)
┌────────────────────┐                  ┌────────────────────┐
│ Kafka Consumer     │                  │ RabbitMQ Consumer  │
│ (Go)               │                  │ (Go)               │
└────────┬───────────┘                  └────────┬───────────┘
         │                                       │
         └───────────────────┬───────────────────┘
                             │ Almacena datos
                             ▼
                      ┌──────────────┐
                      │    Valkey    │
                      │ (Base Datos  │
                      │  en Memoria) │
                      │  2 Réplicas  │
                      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │   Grafana    │
                      │ (Dashboard)  │
                      └──────────────┘
```

### 2.2 Flujo de Datos

1. **Ingesta**: Locust genera 10,000 requests con tweets del clima (Chinautla).
2. **Procesamiento**: API REST en Rust recibe y valida los tweets.
3. **Comunicación**: Go Deployment 1 actúa como servidor gRPC.
4. **Distribución**: Se envía a Kafka y RabbitMQ simultáneamente.
5. **Consumo**: Consumidores separan datos por broker y almacenan en Valkey.
6. **Visualización**: Grafana consulta Valkey y genera dashboards con métricas.

---

## 3. Componentes Detallados

### 3.1 API REST en Rust

**Archivo:** `api-rust/src/main.rs`

**Responsabilidades:**
- Recibir peticiones POST en `POST /api/tweets`
- Validar estructura JSON según proto (municipality, temperature, humidity, weather)
- Convertir enumeraciones de JSON a protobuf
- Enviar vía gRPC a Go Deployment 1
- Escalabilidad mediante HPA (1-3 réplicas)

**Endpoints:**
- `POST /api/tweets` - Recibe un tweet
- `GET /health` - Verificación de salud

**Ejemplo de Request:**
```json
{
  "municipality": "chinautla",
  "temperature": 25,
  "humidity": 65,
  "weather": "sunny"
}
```

**Configuración HPA:**
- Mínimo: 1 réplica
- Máximo: 3 réplicas
- Umbral CPU: > 30%

### 3.2 Go Deployment 1 (gRPC Server + Clients)

**Archivo:** `go-services/deployment1.go`

**Responsabilidades:**
- Servidor gRPC (Puerto 50051) que implementa `WeatherTweetService`
- Cliente de Kafka Writer (envía a Kafka)
- Cliente de RabbitMQ Sender (envía a RabbitMQ)
- API REST interna (Puerto 8081) para health checks

**Servicio gRPC:**
```proto
service WeatherTweetService {
    rpc SendTweet (WeatherTweetRequest) returns (WeatherTweetResponse);
}
```

### 3.3 Message Brokers

#### 3.3.1 Kafka

**Archivo:** `k8s-manifests/06-kafka.yaml`

- **Topic:** `weather-tweets`
- **Particiones:** 3
- **Replicación:** 1
- **Retención:** 168 horas (7 días)
- **Consumer Group:** `weather-consumer-kafka`

**Consumidor:** Go service que lee y almacena en Valkey

#### 3.3.2 RabbitMQ

**Archivo:** `k8s-manifests/07-rabbitmq.yaml`

- **Exchange:** `weather-exchange` (Direct)
- **Queue:** `weather-tweets`
- **Routing Key:** `weather`
- **Usuario:** guest/guest

**Consumidor:** Go service que lee y almacena en Valkey

### 3.4 Valkey (Base de Datos en Memoria)

**Archivo:** `k8s-manifests/05-valkey.yaml`

**Configuración:**
- **StatefulSet:** 2 réplicas
- **Almacenamiento:** 1Gi por réplica
- **Persistencia:** Habilitada (AOF - Append-Only File)
- **Puerto:** 6379

**Estructura de Datos en Valkey:**

```
# Contador de tweets por condición climática
weather:sunny = 2500
weather:cloudy = 2400
weather:rainy = 2600
weather:foggy = 2500

# Datos detallados por municipio y clima
weather:data:chinautla:sunny = {temperature: 25, humidity: 65}
```

### 3.5 Grafana

**Archivo:** `k8s-manifests/09-grafana.yaml`

**Configuración:**
- **Puerto:** 3000
- **Usuario:** admin
- **Contraseña:** admin123

**Dashboard Requerido (Chinautla):**

Gráfica de barras mostrando:
- Total de reportes por condición climática (sunny, cloudy, rainy, foggy)

---

## 4. Manifiestos Kubernetes

### 4.1 Namespace

**Archivo:** `01-namespace.yaml`

Define el namespace `weather-system` donde se despliegan todos los componentes.

### 4.2 Deployments Principales

| Deployment | Archivo | Réplicas | Puerto | Descripción |
|------------|---------|----------|--------|-------------|
| rust-api | `02-rust-api.yaml` | 1-3 (HPA) | 8080 | API REST Rust |
| go-deployment-1 | `03-go-deployment-1.yaml` | 1 | 50051/8081 | Server gRPC + Clients |
| kafka-consumer | `04-consumers.yaml` | 1 | N/A | Consume de Kafka |
| rabbitmq-consumer | `04-consumers.yaml` | 1 | N/A | Consume de RabbitMQ |
| grafana | `09-grafana.yaml` | 1 | 3000 | Dashboard de visualización |

### 4.3 Ingress

**Archivo:** `08-ingress.yaml`

Rutas disponibles:
- `weather.local/api/tweets` → rust-api:8080
- `weather.local/health` → rust-api:8080
- `grafana.local` → grafana:3000

---

## 5. Instrucciones de Despliegue

### 5.1 Requisitos Previos

1. **Clúster GKE activo** en proyecto-3-475405
2. **Credenciales configuradas:**
   ```bash
   gcloud container clusters get-credentials proyecto3-sopes-1 --zone us-central1-c --project proyecto-3-475405
   ```
3. **Kubectl configurado y funcionando**
4. **NGINX Ingress Controller instalado** (si no está):
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/gcp/deploy.yaml
   ```

### 5.2 Pasos de Despliegue

```bash
# 1. Crear namespace
kubectl apply -f proyecto3/k8s-manifests/01-namespace.yaml

# 2. Desplegar base de datos y message brokers
kubectl apply -f proyecto3/k8s-manifests/05-valkey.yaml
kubectl apply -f proyecto3/k8s-manifests/06-kafka.yaml
kubectl apply -f proyecto3/k8s-manifests/07-rabbitmq.yaml

# Esperar a que estén listos (5-10 minutos)
kubectl get pods -n weather-system

# 3. Desplegar servicios principales (requiere imágenes en Zot)
kubectl apply -f proyecto3/k8s-manifests/02-rust-api.yaml
kubectl apply -f proyecto3/k8s-manifests/03-go-deployment-1.yaml
kubectl apply -f proyecto3/k8s-manifests/04-consumers.yaml

# 4. Desplegar visualización
kubectl apply -f proyecto3/k8s-manifests/09-grafana.yaml

# 5. Desplegar Ingress
kubectl apply -f proyecto3/k8s-manifests/08-ingress.yaml

# Verificar status
kubectl get all -n weather-system
kubectl get ingress -n weather-system
```

### 5.3 Construcción y Push de Imágenes a Zot

```bash
# Reemplazar ZOT_REGISTRY con IP de VM (ej: 34.159.50.100:5000)

# API Rust
cd proyecto3/api-rust
docker build -t ZOT_REGISTRY/weather-api-rust:latest .
docker push ZOT_REGISTRY/weather-api-rust:latest

# Go Services
cd ../go-services
docker build -f Dockerfile.deployment1 -t ZOT_REGISTRY/go-services-deployment1:latest .
docker push ZOT_REGISTRY/go-services-deployment1:latest

docker build -f Dockerfile.kafka-consumer -t ZOT_REGISTRY/go-kafka-consumer:latest .
docker push ZOT_REGISTRY/go-kafka-consumer:latest

docker build -f Dockerfile.rabbitmq-consumer -t ZOT_REGISTRY/go-rabbitmq-consumer:latest .
docker push ZOT_REGISTRY/go-rabbitmq-consumer:latest
```

---

## 6. Pruebas de Carga con Locust

### 6.1 Instalación

```bash
cd proyecto3/locust-config
pip install -r requirements.txt
```

### 6.2 Ejecución

```bash
# Modo interactivo
locust -f locustfile.py --host=http://INGRESS_IP

# Modo sin interfaz (10,000 requests, 10 usuarios concurrentes)
locust -f locustfile.py \
  --host=http://INGRESS_IP \
  --headless \
  -u 10 \
  -r 10 \
  -n 10000
```

### 6.3 Monitoreo

Durante la carga:
1. Observar HPA del Rust API: `kubectl get hpa -n weather-system -w`
2. Verificar pods escalados: `kubectl get pods -n weather-system`
3. Ver métricas en Grafana: `http://grafana.local`

---

## 7. Análisis de Rendimiento

### 7.1 Kafka vs RabbitMQ

| Aspecto | Kafka | RabbitMQ |
|---------|-------|----------|
| **Throughput** | Mayor (100K+ msg/s) | Moderado (50K msg/s) |
| **Latencia** | Más alta (~100ms) | Más baja (~10ms) |
| **Persistencia** | Nativa, Durable | Configurable |
| **Escalabilidad** | Horizontal (particiones) | Vertical (clustering) |
| **Replicación** | Multi-broker | Mirroring |

### 7.2 Impacto de Réplicas en Valkey

**Con 1 Réplica:**
- Mejor rendimiento: ~10K ops/sec
- Riesgo: Pérdida de datos en caso de fallo
- Almacenamiento: Reducido

**Con 2 Réplicas:**
- Rendimiento: ~8K ops/sec (por sincronización)
- Confiabilidad: Alta (tolerancia a 1 fallo)
- Almacenamiento: Duplicado

**Recomendación:** 2 réplicas para producción

### 7.3 REST vs gRPC

| Aspecto | REST (Rust) | gRPC (Go) |
|---------|-------------|-----------|
| **Protocolo** | HTTP/1.1 | HTTP/2 |
| **Serialización** | JSON | Protocol Buffers |
| **Tamaño** | Mayor (~500 bytes) | Menor (~200 bytes) |
| **Latencia** | ~50ms | ~10ms |
| **Tooling** | Amplio | Especializado |

---

## 8. Respuestas a Preguntas Técnicas

### 8.1 ¿Cómo funciona Kafka en el proyecto?

Kafka recibe mensajes de clima desde Go Deployment 1 y los almacena en el tópico `weather-tweets`. El consumidor de Kafka lee estos mensajes en grupos, los procesa y almacena contadores en Valkey por tipo de clima.

**Ventaja:** Mejor para alto volumen y análisis histórico.

### 8.2 ¿Diferencia entre Valkey y RabbitMQ?

- **Valkey:** Base de datos en memoria, almacena datos procesados (contadores)
- **RabbitMQ:** Message broker, transmite mensajes entre servicios

Son complementarios: RabbitMQ entrega, Valkey almacena.

### 8.3 ¿Por qué usar gRPC?

- Serialización eficiente (Protocol Buffers)
- HTTP/2: Multiplexing y menor latencia
- Fuerte tipado mediante `.proto`
- Mejor rendimiento para comunicación inter-servicio

### 8.4 ¿Cómo implementar HPA?

Ya está configurado en `02-rust-api.yaml`:
```yaml
metrics:
- type: Resource
  resource:
    name: cpu
    target:
      type: Utilization
      averageUtilization: 30
```

Se escala cuando CPU > 30% en promedio.

### 8.5 ¿Cómo mejorar con replicación?

**Valkey:** StatefulSet con 2 réplicas proporciona:
- Persistencia: AOF (Append-Only File)
- Tolerancia a fallos: Una réplica puede fallar
- Backup automático

**Mejora adicional:**
```yaml
# Agregar VPA para optimizar recursos
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: valkey-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: StatefulSet
    name: valkey
  updatePolicy:
    updateMode: "Auto"
```

---

## 9. Conclusiones y Lecciones Aprendidas

### 9.1 Arquitectura Escalable

El sistema demuestra:
- **Separación de responsabilidades:** Cada componente tiene un rol claro
- **Escalabilidad horizontal:** HPA en Rust API, particiones en Kafka
- **Redundancia:** Múltiples réplicas de Valkey

### 9.2 Elección de Tecnologías

- **Rust:** API REST rápida y segura
- **Go:** Ideal para servicios concurrentes y conectores
- **Kafka/RabbitMQ:** Diferentes casos de uso
- **Valkey:** Almacenamiento rápido de datos procesados
- **Grafana:** Visualización en tiempo real

### 9.3 Mejoras Futuras

1. **Autenticación y Autorización:** OAuth2/OIDC
2. **Encryption:** TLS en comunicaciones, secrets en K8s
3. **Monitoreo:** Prometheus + AlertManager
4. **Resiliencia:** Circuit Breakers, Retries
5. **Rate Limiting:** Para proteger API REST

---

## 10. Referencias

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Rust Actix-web](https://actix.rs/)
- [Go gRPC](https://grpc.io/docs/languages/go/)
- [Kafka Documentation](https://kafka.apache.org/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Valkey Documentation](https://valkey.io/)
- [Grafana Documentation](https://grafana.com/docs/)

---

## Anexos

### A. Configuración Local (Docker Compose)

Para pruebas locales sin Kubernetes, consultar `docker-compose.yml` (opcional).

### B. Scripts de Utilidad

**Ver logs de todos los pods:**
```bash
kubectl logs -n weather-system -f deployment/go-deployment-1
```

**Escalar manualmente (antes de demostración):**
```bash
kubectl scale deployment rust-api -n weather-system --replicas=3
```

**Acceder a Valkey:**
```bash
kubectl exec -it valkey-0 -n weather-system -- redis-cli
```

---

**Documento preparado por:** Luis Pablo Manuel García López  
**Carnet:** 202200129  
**Fecha:** Octubre 18, 2025
