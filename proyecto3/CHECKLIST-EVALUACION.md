# ✅ CHECKLIST DE EVALUACIÓN - Proyecto 3 SOPES

**Carnet:** 202200129  
**Municipio:** Chinautla  
**Fecha:** 18 Octubre 2025  
**Evaluación:** 23-25 Octubre 2025

---

## 📋 REQUISITOS DEL ENUNCIADO

### 1. ARQUITECTURA DE MICROSERVICIOS ✓

- [ ] Sistema distribuido con múltiples componentes
- [ ] Comunicación via gRPC (entre API y Go Service)
- [ ] Comunicación via REST (Locust → API)
- [ ] Componentes independientes y escalables
  - [ ] API Rust (Deployment con HPA)
  - [ ] Go Service (gRPC server)
  - [ ] Kafka Consumer
  - [ ] RabbitMQ Consumer
  - [ ] Valkey (Redis)

**Ubicación:** `/api-rust/`, `/go-services/`, `k8s-manifests/`

---

### 2. BASE DE DATOS - VALKEY (Redis)

- [ ] Valkey/Redis funcionando en cluster
- [ ] Datos persistidos (AOF enabled)
- [ ] Replicación (2 replicas)
- [ ] Contadores de clima:
  - [ ] `weather:sunny`
  - [ ] `weather:cloudy`
  - [ ] `weather:rainy`
  - [ ] `weather:foggy`
- [ ] Datos detallados por municipio:
  - [ ] `weather:data:chinautla:sunny`
  - [ ] `weather:data:chinautla:cloudy`
  - [ ] `weather:data:chinautla:rainy`
  - [ ] `weather:data:chinautla:foggy`

**Prueba:**
```bash
./run-valkey.sh
# Opción 3: Ver contadores
```

**Archivo:** `k8s-manifests/05-valkey.yaml`

---

### 3. SISTEMA DE COLAS DE MENSAJES

#### KAFKA ✓

- [ ] Cluster Kafka desplegado
- [ ] Topic `weather-tweets` creado
- [ ] Particiones: 3
- [ ] Replicación factor: 1
- [ ] Kafka Consumer leyendo y almacenando en Valkey

**Prueba:**
```bash
kubectl exec -it kafka-0 -n weather-system -- \
  kafka-topics --list --bootstrap-server localhost:9092
```

**Archivo:** `k8s-manifests/06-kafka.yaml`

#### RabbitMQ ✓

- [ ] RabbitMQ desplegado
- [ ] Exchange `weather-exchange` creado
- [ ] Cola `weather-tweets` creada
- [ ] RabbitMQ Consumer leyendo y almacenando en Valkey

**Prueba:**
```bash
kubectl port-forward svc/rabbitmq 15672:15672 -n weather-system &
# Acceder a http://localhost:15672
# Usuario: guest, Contraseña: guest
```

**Archivo:** `k8s-manifests/07-rabbitmq.yaml`

---

### 4. AUTOSCALADO (HPA)

- [ ] Horizontal Pod Autoscaler configurado
- [ ] Target: Deployment `rust-api`
- [ ] Métrica: CPU utilization
- [ ] Threshold: 30%
- [ ] Min replicas: 1
- [ ] Max replicas: 3
- [ ] Escalado funciona bajo carga

**Prueba:**
```bash
kubectl get hpa -n weather-system -w
# Observar que REPLICAS sube de 1 a 3 mientras Locust corre
```

**Archivo:** `k8s-manifests/02-rust-api.yaml`

---

### 5. PROTOCOLO BUFFERS (gRPC)

- [ ] Archivo `weather_tweet.proto` definido
- [ ] Mensaje `WeatherTweetRequest`:
  - [ ] `municipality` (string)
  - [ ] `temperature` (float)
  - [ ] `humidity` (float)
  - [ ] `weather` (enum)
- [ ] Mensaje `WeatherTweetResponse`:
  - [ ] `status` (string)
  - [ ] `message` (string)
- [ ] Service `WeatherTweetService` con RPC `SendTweet`

**Archivo:** `proto/weather_tweet.proto`

---

### 6. CONTENEDORES DOCKER

- [ ] 4 imágenes compiladas exitosamente:
  - [ ] `weather-api-rust`
  - [ ] `go-services-deployment1`
  - [ ] `go-kafka-consumer`
  - [ ] `go-rabbitmq-consumer`
- [ ] Imágenes almacenadas en Zot Registry
- [ ] Imágenes usadas en Kubernetes manifests

**Prueba:**
```bash
./build-images.sh 34.159.50.100:5000
curl http://34.159.50.100:5000/v2/_catalog
```

---

### 7. ORQUESTACIÓN KUBERNETES

- [ ] Namespace `weather-system` creado
- [ ] 9 manifests YAML numerados y en orden:
  - [ ] 01-namespace.yaml
  - [ ] 02-rust-api.yaml (Deployment + HPA + Service)
  - [ ] 03-go-deployment-1.yaml (Deployment + Service)
  - [ ] 04-consumers.yaml (Kafka + RabbitMQ consumers)
  - [ ] 05-valkey.yaml (StatefulSet)
  - [ ] 06-kafka.yaml (Zookeeper + Kafka)
  - [ ] 07-rabbitmq.yaml (RabbitMQ)
  - [ ] 08-ingress.yaml (NGINX Ingress)
  - [ ] 09-grafana.yaml (Grafana dashboard)
- [ ] Todos los pods en estado `Running`
- [ ] Services con endpoints correctos
- [ ] Ingress Controller asignando IP

**Prueba:**
```bash
./deploy.sh
kubectl get all -n weather-system
```

**Archivos:** `k8s-manifests/*.yaml`

---

### 8. GENERADOR DE CARGA - LOCUST

- [ ] Script Locust funcional (`locustfile.py`)
- [ ] Cliente simula usuarios concurrentes
- [ ] Genera requests a `/api/tweets`
- [ ] Datos contienen:
  - [ ] Municipio: Chinautla
  - [ ] Temperatura aleatoria (15-35°C)
  - [ ] Humedad aleatoria (30-90%)
  - [ ] Clima aleatorio (sunny/cloudy/rainy/foggy)
- [ ] Capaz de generar 10,000+ tweets
- [ ] Puede ejecutarse en modo web e headless

**Prueba:**
```bash
./run-locust.sh
# Opción 4: 10,000 tweets
```

**Archivo:** `locust-config/locustfile.py`

---

### 9. VISUALIZACIÓN - GRAFANA

- [ ] Grafana desplegado en Kubernetes
- [ ] Accesible en `http://localhost:3000`
- [ ] Credenciales: admin / admin123
- [ ] Data Source Redis configurado (redis://valkey:6379)
- [ ] Dashboard con paneles mostrando:
  - [ ] Contadores de clima en tiempo real
  - [ ] Temperatura promedio por clima
  - [ ] Humedad promedio por clima
- [ ] Datos se actualizan mientras Locust corre

**Prueba:**
```bash
./run-grafana.sh
# Opción 1: Abrir en navegador
```

**Archivo:** `k8s-manifests/09-grafana.yaml`

---

## 🎯 REQUISITOS FUNCIONALES

### Flujo Completo ✓

```
┌──────────────┐
│   LOCUST     │  Genera 10,000 tweets
│  (Cliente)   │
└──────┬───────┘
       │ HTTP POST /api/tweets
       ▼
┌──────────────┐
│ RUST API     │  Valida y convierte a protobuf
│  :8080       │
└──────┬───────┘
       │ gRPC :50051
       ▼
┌──────────────────┐
│ GO SERVICE       │
│ :50051 gRPC      │
│ :8081 REST       │
└──────┬───────────┘
   ┌───┴────┐
   │        │
   ▼        ▼
┌──────┐ ┌────────┐
│KAFKA │ │RabbitMQ│  Ambos
└───┬──┘ └──┬─────┘  simultáneamente
    │       │
    ▼       ▼
┌──────────────────┐
│ CONSUMIDORES     │
│ (Ambos)          │
└────────┬─────────┘
         │
         ▼
   ┌──────────────┐
   │ VALKEY       │  Almacena contadores
   │ :6379        │
   └──────────────┘
         △
         │
    ┌────┴──────┐
    │            │
    ▼            ▼
┌─────────┐  ┌────────┐
│GRAFANA  │  │redis-cli│ Lectura
│:3000    │  └────────┘
└─────────┘
```

- [ ] Locust → Rust API ✓
- [ ] Rust API → Go Service (gRPC) ✓
- [ ] Go Service → Kafka ✓
- [ ] Go Service → RabbitMQ ✓
- [ ] Kafka Consumer → Valkey ✓
- [ ] RabbitMQ Consumer → Valkey ✓
- [ ] Valkey → Grafana ✓

---

### Pruebas Bajo Carga ✓

- [ ] Sistema maneja 10,000 tweets sin errores
- [ ] API Rust responde < 200ms por request
- [ ] HPA escala a 3 replicas bajo carga
- [ ] Kafka y RabbitMQ procesan sin pérdida de mensajes
- [ ] Valkey almacena sin pérdida de datos
- [ ] Grafana muestra datos en tiempo real

---

## 📊 MÉTRICAS A CAPTURAR

### Antes de Locust

```bash
./run-valkey.sh
# Opción 3

# Documento: Contadores iniciales = 0
```

### Durante Locust

```bash
# Terminal 1: Locust corriendo
./run-locust.sh

# Terminal 2: Monitorear HPA
kubectl get hpa -n weather-system -w

# Terminal 3: Ver contadores actualizando
watch -n 1 'kubectl exec valkey-0 -n weather-system -- redis-cli GET weather:sunny'
```

### Después de Locust

```bash
./run-valkey.sh
# Opción 3

# Documento: Contadores finales (debe ser > 10,000 entre todos)
# Ejemplo:
# weather:sunny: 2,450
# weather:cloudy: 1,890
# weather:rainy: 3,120
# weather:foggy: 2,540
```

---

## 📸 CAPTURAS REQUERIDAS

### Para Calificación, Captura:

1. **Cluster Running**
   ```bash
   gcloud container clusters list
   # Mostrar proyecto3-sopes-1 Running
   ```

2. **Pods Desplegados**
   ```bash
   kubectl get pods -n weather-system
   # Mostrar todos en Running
   ```

3. **HPA Escalando**
   ```bash
   kubectl get hpa -n weather-system
   # Mostrar REPLICAS = 3, TARGETS > 30%
   ```

4. **Grafana Dashboard**
   - Acceso en navegador
   - Mostrar contadores en tiempo real
   - Mostrar datos actualizándose

5. **Valkey Contadores**
   ```bash
   ./run-valkey.sh
   # Opción 3
   # Mostrar valores > 0
   ```

6. **Locust Output**
   - Requests/segundo
   - Latencia promedio
   - Total de requests generados

7. **Logs de Servicios**
   ```bash
   kubectl logs deployment/go-deployment1 -n weather-system
   # Mostrar procesamiento de mensajes
   ```

---

## ⚠️ ADVERTENCIAS IMPORTANTES

### NO HACER ANTES DE EVALUACIÓN

- [ ] ❌ No destruir el cluster
- [ ] ❌ No hacer `cleanup.sh` si aún no termina evaluación
- [ ] ❌ No vaciar Valkey con `FLUSHALL`
- [ ] ❌ No modificar los manifests después de desplegar
- [ ] ❌ No reducir réplicas de Valkey

### SÍ HACER ANTES DE EVALUACIÓN

- [ ] ✅ Documentar todas las métricas
- [ ] ✅ Capturar screenshots de todo
- [ ] ✅ Anotar IPs importantes
- [ ] ✅ Revisar logs para errores
- [ ] ✅ Probar Grafana y Valkey queries

---

## 🚀 ÚLTIMA VERIFICACIÓN (Día de Evaluación)

```bash
# 1. Verificar cluster sigue corriendo
gcloud container clusters describe proyecto3-sopes-1 \
  --zone us-central1-c --project proyecto-3-475405 | grep status

# 2. Verificar todos los pods
kubectl get pods -n weather-system

# 3. Verificar HPA
kubectl get hpa -n weather-system

# 4. Verificar Ingress
kubectl get ingress -n weather-system

# 5. Verificar datos en Valkey
./run-valkey.sh
# Opción 3

# 6. Verificar Grafana
./run-grafana.sh
# Opción 1

# 7. Mostrar logs recientes
kubectl logs deployment/go-deployment-1 -n weather-system --tail=50
```

---

## ✓ CHECKLIST FINAL

- [ ] Cluster desplegado hace 24+ horas
- [ ] Todos los pods en estado `Running`
- [ ] HPA funciona (replica count = 3 bajo carga)
- [ ] Kafka y RabbitMQ procesan mensajes
- [ ] Valkey contiene datos
- [ ] Grafana visualiza datos
- [ ] Locust puede generar 10,000 tweets
- [ ] Documentación completa en 5 archivos markdown
- [ ] Scripts funcionan sin errores
- [ ] Evidencia capturada en screenshots

---

## 📞 INFORMACIÓN DEL ESTUDIANTE

**Carnet:** 202200129  
**Nombre:** (Tu nombre)  
**Municipio:** Chinautla  
**Universidad:** USAC  
**Curso:** SO1  
**Sección:** (Tu sección)  
**Proyecto:** 3  
**Año:** 2025  

---

**Fecha de Despliegue:** 18 Octubre 2025  
**Fecha de Evaluación:** 23-25 Octubre 2025  
**Cluster Uptime Requerido:** 24+ horas

---

*Última actualización: 18 Octubre 2025*
