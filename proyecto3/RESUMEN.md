# 📊 PROYECTO 3 - SISTEMA DE TWEETS DEL CLIMA

## 📋 Información del Proyecto

| Item | Valor |
|------|-------|
| **Carnet** | 202200129 |
| **Estudiante** | Luis Pablo Manuel García López |
| **Municipio** | Chinautla (carnet termina en 9) |
| **Fecha** | Octubre 18, 2025 |
| **Estado** | ✅ Completado |

---

## 🎯 Resumen Ejecutivo

Sistema distribuido en **Google Kubernetes Engine (GKE)** que procesa "tweets" sobre el clima local en tiempo real. Incluye:

- ✅ **API REST en Rust** con escalabilidad automática (HPA)
- ✅ **Servicios en Go** con comunicación gRPC
- ✅ **Message Brokers**: Kafka y RabbitMQ
- ✅ **Base de datos**: Valkey (en memoria, 2 réplicas)
- ✅ **Visualización**: Grafana con dashboard
- ✅ **Generador de carga**: Locust (10,000 requests)
- ✅ **Ingress Controller**: NGINX
- ✅ **Registry**: Zot (en VM GCP)

---

## 📁 Estructura del Proyecto

```
proyecto3/
├── 📂 proto/                   # Protocol Buffers
│   └── weather_tweet.proto
├── 📂 api-rust/               # API REST (Rust)
│   ├── src/main.rs
│   ├── Cargo.toml
│   ├── build.rs
│   └── Dockerfile
├── 📂 go-services/            # Servicios (Go)
│   ├── deployment1.go (Server gRPC + Clients)
│   ├── kafka_writer.go
│   ├── kafka_consumer.go
│   ├── rabbitmq_sender.go
│   ├── rabbitmq_consumer.go
│   └── Dockerfiles
├── 📂 k8s-manifests/          # Kubernetes YAMLs
│   ├── 01-namespace.yaml
│   ├── 02-rust-api.yaml (con HPA)
│   ├── 03-go-deployment-1.yaml
│   ├── 04-consumers.yaml
│   ├── 05-valkey.yaml (2 réplicas)
│   ├── 06-kafka.yaml
│   ├── 07-rabbitmq.yaml
│   ├── 08-ingress.yaml
│   └── 09-grafana.yaml
├── 📂 locust-config/          # Pruebas de carga
│   ├── locustfile.py
│   └── requirements.txt
├── 🔧 Scripts de Automatización
│   ├── deploy.sh
│   ├── cleanup.sh
│   ├── build-images.sh
│   ├── update-registry.sh
│   └── monitor.sh
├── 📚 Documentación
│   ├── TECNICA.md (Documentación técnica completa)
│   ├── README.md (Guía de inicio rápido)
│   ├── EJEMPLOS.md (Ejemplos de uso)
│   ├── GRAFANA-SETUP.md (Configuración del dashboard)
│   └── RESUMEN.md (Este archivo)
└── .gitignore

```

---

## 🚀 Inicio Rápido (5 minutos)

### 1️⃣ Configurar GCP
```bash
gcloud container clusters get-credentials proyecto3-sopes-1 \
  --zone us-central1-c --project proyecto-3-475405
```

### 2️⃣ Actualizar Zot Registry
```bash
cd proyecto3
./update-registry.sh 34.159.50.100:5000  # Reemplazar IP
```

### 3️⃣ Compilar imágenes Docker
```bash
./build-images.sh 34.159.50.100:5000
```

### 4️⃣ Desplegar sistema
```bash
./deploy.sh
```

### 5️⃣ Ejecutar pruebas
```bash
cd locust-config
locust -f locustfile.py --host=http://INGRESS_IP
```

---

## 🏗️ Arquitectura del Sistema

```
┌──────────────┐
│   Locust     │ Genera 10K tweets
│  (Cliente)   │
└──────┬───────┘
       │ HTTP
       ▼
┌──────────────────────────────────────────────────────────────┐
│                 INGRESS NGINX CONTROLLER                     │
└──────────────────────┬─────────────────────────────────────┘
                       │
       ┌───────────────┴───────────────┐
       ▼ (/api/tweets)                 ▼ (/)
   ┌─────────────┐              ┌────────────┐
   │ API Rust    │ HPA: 1-3     │  Grafana   │
   │ Port 8080   │ CPU > 30%    │ Port 3000  │
   └─────┬───────┘              └────────────┘
         │ gRPC
         ▼
   ┌──────────────────────────────────────────────┐
   │    Go Deployment 1 - gRPC Server             │
   │    - Recibe requests de Rust                 │
   │    - Publica en Kafka                        │
   │    - Publica en RabbitMQ                     │
   │    Port 50051 (gRPC), 8081 (REST)            │
   └──────────┬──────────────────────────┬────────┘
              │                          │
              ▼ (Kafka Topic)            ▼ (RabbitMQ Queue)
        ┌────────────┐             ┌─────────────┐
        │   Kafka    │             │  RabbitMQ   │
        │  Broker    │             │   Broker    │
        └──────┬─────┘             └──────┬──────┘
               │                         │
               ▼ (Consumer)              ▼ (Consumer)
        ┌──────────────┐         ┌────────────────┐
        │ Kafka        │         │ RabbitMQ       │
        │ Consumer     │         │ Consumer       │
        │ (Go)         │         │ (Go)           │
        └──────┬───────┘         └────────┬───────┘
               └─────────────┬────────────┘
                             │ Almacena
                             ▼
                    ┌──────────────────┐
                    │  Valkey (Redis)  │
                    │  2 Réplicas      │
                    │  Persistencia:ON │
                    │  Port 6379       │
                    └────────┬─────────┘
                             │ Lee datos
                             ▼
                    ┌──────────────────┐
                    │  Grafana         │
                    │  Dashboard       │
                    │  (Visualiza)     │
                    └──────────────────┘
```

---

## 📊 Flujo de Datos

### 1. **Ingesta de Datos**
   - Locust genera tweets: `{municipality, temperature, humidity, weather}`
   - POST → API REST Rust

### 2. **Validación y Conversión**
   - API Rust valida estructura JSON
   - Convierte a Protocol Buffers
   - Envía vía gRPC a Go Deployment 1

### 3. **Distribución de Mensajes**
   - Go Deployment 1 recibe en servidor gRPC
   - Publica simultáneamente en:
     - **Kafka Topic**: `weather-tweets`
     - **RabbitMQ Queue**: `weather-tweets`

### 4. **Consumo y Almacenamiento**
   - **Kafka Consumer**: Lee → Incrementa contador en Valkey
   - **RabbitMQ Consumer**: Lee → Incrementa contador en Valkey

### 5. **Visualización**
   - Grafana consulta Valkey
   - Dashboard muestra gráfico de barras:
     - **X-axis**: Condiciones climáticas (sunny, cloudy, rainy, foggy)
     - **Y-axis**: Total de reportes

---

## 🔑 Características Principales

### ✅ Completados

- [x] **Namespace**: `weather-system`
- [x] **API REST Rust**: Recibe y valida tweets
- [x] **HPA**: Escalabilidad automática (1-3 réplicas)
- [x] **Go Services**: 
  - [x] Deployment 1: gRPC Server + Clients
  - [x] Kafka Writer
  - [x] Kafka Consumer
  - [x] RabbitMQ Sender
  - [x] RabbitMQ Consumer
- [x] **Message Brokers**:
  - [x] Kafka (Strimzi): Topic `weather-tweets`
  - [x] RabbitMQ: Queue `weather-tweets`
- [x] **Valkey**: 
  - [x] StatefulSet 2 réplicas
  - [x] Persistencia AOF
- [x] **Grafana**: Dashboard requerido
- [x] **Ingress**: NGINX Controller
- [x] **Locust**: Generador de carga
- [x] **Zot**: Container Registry
- [x] **Documentación**: Técnica completa
- [x] **Scripts**: Despliegue y monitoreo

---

## 📈 Requisitos Cumplidos del Enunciado

| Requisito | Status | Detalles |
|-----------|--------|----------|
| Locust + JSON | ✅ | 10,000 tweets a Ingress |
| API Rust + HPA | ✅ | 1-3 réplicas, CPU > 30% |
| Go Deployment 1 | ✅ | gRPC + Kafka + RabbitMQ |
| Kafka | ✅ | Topic weather-tweets |
| RabbitMQ | ✅ | Queue weather-tweets |
| Consumidores | ✅ | Kafka y RabbitMQ |
| Valkey | ✅ | 2 réplicas, persistencia |
| Grafana | ✅ | Dashboard de barras |
| Ingress | ✅ | NGINX Controller |
| Zot Registry | ✅ | Imágenes alojadas |
| OCI Artifact | ⚠️ | Documentado en TECNICA.md |
| Documentación | ✅ | TECNICA.md completo |
| Namespaces | ✅ | weather-system |

---

## 🛠️ Tecnologías Utilizadas

| Componente | Tecnología | Versión |
|------------|------------|---------|
| **Cloud** | Google Cloud Platform | - |
| **Orquestación** | Kubernetes (GKE) | 1.24+ |
| **Ingress** | NGINX Ingress Controller | 1.8+ |
| **API** | Rust (Actix-web) | 2021 edition |
| **Services** | Go | 1.21 |
| **Serialización** | Protocol Buffers 3 | - |
| **Broker 1** | Apache Kafka | 7.5.0 |
| **Broker 2** | RabbitMQ | 3.12 |
| **BD en Memoria** | Valkey | 7 |
| **Visualización** | Grafana | Latest |
| **Pruebas** | Locust | 2.17+ |
| **Registry** | Zot | - |

---

## 📚 Documentación

| Archivo | Propósito |
|---------|-----------|
| **TECNICA.md** | 📖 Documentación técnica completa |
| **README.md** | 🚀 Inicio rápido |
| **EJEMPLOS.md** | 📝 Ejemplos de uso |
| **GRAFANA-SETUP.md** | 📊 Configuración de dashboard |
| **RESUMEN.md** | 📋 Este archivo |

---

## ⚡ Comandos Útiles

```bash
# Ver estado
kubectl get all -n weather-system

# Ver HPA
kubectl get hpa -n weather-system -w

# Ver logs
kubectl logs -f deployment/go-deployment-1 -n weather-system

# Monitoreo
./monitor.sh

# Escalar manualmente
kubectl scale deployment rust-api -n weather-system --replicas=3

# Acceder a Valkey
kubectl exec -it valkey-0 -n weather-system -- redis-cli

# Limpiar todo
./cleanup.sh
```

---

## 🎓 Aprendizajes Clave

1. **Microservicios**: Separación de responsabilidades
2. **Kubernetes**: Deployments, Services, HPA, StatefulSets
3. **gRPC**: Comunicación eficiente entre servicios
4. **Message Brokers**: Kafka vs RabbitMQ
5. **Escalabilidad**: Horizontal Pod Autoscaler
6. **Persistencia**: AOF en bases de datos en memoria
7. **Visualización**: Grafana como herramienta de BI

---

## ⚠️ Notas Importantes

1. **Para calificación presencial**:
   - Cluster debe estar corriendo 24 horas antes ✅
   - Valkey debe estar **VACÍO** al inicio ✅
   - Todas las imágenes en Zot Registry ✅
   - Pestañas del navegador listas ✅

2. **Municipio**: Chinautla (carnet 202200129, último dígito 9)

3. **Dashboard**: Gráfica de barras con conteos por clima

4. **Sin prórroga**: Fecha límite: 22 de octubre de 2025

---

## 📞 Soporte

Para problemas:
1. Ver TECNICA.md → Sección de Troubleshooting
2. Ver EJEMPLOS.md → Sección de Troubleshooting
3. Revisar logs: `kubectl logs <pod> -n weather-system`
4. Contactar: Luis Pablo García López (202200129)

---

**Estado del Proyecto:** ✅ **COMPLETADO Y LISTO PARA DESPLIEGUE**

Última actualización: Octubre 18, 2025

