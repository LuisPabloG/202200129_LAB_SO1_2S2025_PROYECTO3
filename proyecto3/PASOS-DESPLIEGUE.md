# 📋 INSTRUCCIONES PASO A PASO - Despliegue y Pruebas

**Carnet:** 202200129  
**Municipio:** Chinautla  
**Curso:** SO1  
**Proyecto:** 3  
**Año:** 2025

---

## ✅ VERIFICACIÓN PREVIA

Antes de comenzar, asegúrate de tener:

- [ ] Cloud Shell abierto en GCP
- [ ] Kubeconfig cargado (`gcloud container clusters get-credentials ...`)
- [ ] IP de tu Zot Registry (ej: `34.159.50.100`)
- [ ] Zot Registry corriendo en tu VM
- [ ] Docker instalado localmente (para Locust)
- [ ] Python 3.7+ instalado localmente

---

## 🚀 PASO 1: ACTUALIZAR IP DE ZOT REGISTRY

Este es el **paso más importante**. El Zot Registry contiene tus imágenes Docker compiladas.

### 1a. Encontrar tu IP de Zot Registry

```bash
# En tu VM donde corre Zot, ejecuta:
hostname -I

# O en Cloud Shell, lista las VMs:
gcloud compute instances list

# Copia la IP externa (ej: 34.159.50.100)
```

### 1b. Actualizar la IP en los manifests

```bash
cd /home/luis-pablo-garcia/Escritorio/PROYECTO\ 3\ SOPES/202200129_LAB_SO1_2S2025_PROYECTO3/proyecto3

# Reemplaza 34.159.50.100 con tu IP real
./update-registry.sh 34.159.50.100:5000

# Verifica que se actualizó
grep -r "34.159.50.100" k8s-manifests/
```

**Salida esperada:** Todos los manifests con tu IP de Zot.

---

## 🐳 PASO 2: COMPILAR IMÁGENES DOCKER

Este paso compila 4 imágenes Docker y las envía a Zot Registry.

```bash
# Compilar y enviar imágenes (5-10 minutos)
./build-images.sh 34.159.50.100:5000
```

**¿Qué hace?**
1. Compila `weather-api-rust` desde `api-rust/`
2. Compila `go-services-deployment1` desde `go-services/`
3. Compila `go-kafka-consumer` desde `go-services/`
4. Compila `go-rabbitmq-consumer` desde `go-services/`
5. Envía todas a Zot Registry

**Salida esperada:**
```
✓ Building weather-api-rust...
✓ Pushing to registry...
✓ Building go-services-deployment1...
...
✓ Todos los builds completados exitosamente
```

**Si falla:**
- Verifica que Zot Registry esté corriendo: `curl http://34.159.50.100:5000/v2/`
- Verifica conectividad: `ping 34.159.50.100`
- Revisa logs del build: `docker logs`

---

## ☸️ PASO 3: DESPLEGAR EN KUBERNETES

Este paso despliega toda la infraestructura en el cluster.

### 3a. Ejecutar el deployment

```bash
# Desplegar (15-20 minutos)
./deploy.sh
```

**Secuencia de despliegue:**
1. Crea namespace `weather-system`
2. Despliega Valkey (5 min para estar listo)
3. Despliega Kafka y Zookeeper (8 min)
4. Despliega RabbitMQ (5 min)
5. Despliega servicios Go y Rust
6. Despliega Grafana
7. Despliega Ingress Controller
8. Espera a que todos estén en estado `Running`

### 3b. Monitorear el despliegue (EN OTRA TERMINAL)

```bash
# Ver pods en vivo
kubectl get pods -n weather-system -w

# Presiona Ctrl+C cuando todos estén en Running o Completed
```

**Salida esperada:**
```
NAME                                 READY   STATUS    RESTARTS   AGE
valkey-0                             1/1     Running   0          5m
zookeeper-0                          1/1     Running   0          8m
kafka-0                              1/1     Running   0          8m
rabbitmq-0                           1/1     Running   0          5m
go-deployment-1-xxxxx                1/1     Running   0          2m
rust-api-xxxxx                       1/1     Running   0          2m
kafka-consumer-xxxxx                 1/1     Running   0          2m
rabbitmq-consumer-xxxxx              1/1     Running   0          2m
grafana-xxxxx                        1/1     Running   0          1m
```

---

## 📍 PASO 4: OBTENER IP DEL INGRESS

```bash
# Esperar y obtener IP
kubectl get ingress -n weather-system

# Salida esperada:
# NAME              CLASS   HOSTS                  ADDRESS         PORTS
# weather-ingress   nginx   weather.local,grafana.local   35.201.XX.XX    80
```

**Copia la IP** (ej: `35.201.123.45`) para los siguientes pasos.

---

## 🔥 PASO 5: EJECUTAR LOCUST (PRUEBAS DE CARGA)

Ahora vamos a generar tweets usando Locust.

### 5a. Ejecutar desde el proyecto

```bash
cd /home/luis-pablo-garcia/Escritorio/PROYECTO\ 3\ SOPES/202200129_LAB_SO1_2S2025_PROYECTO3/proyecto3

./run-locust.sh
```

### 5b. Seleccionar opción de carga

```
════ OPCIONES DE CARGA ════

1. Modo Web (Interface Interactiva) - RECOMENDADO
2. Modo Headless Ligero (1000 tweets, 5 usuarios)
3. Modo Headless Medio (5000 tweets, 10 usuarios)
4. Modo Headless Pesado (10000 tweets, 20 usuarios)
5. Personalizado

Selecciona una opción (1-5): 4
```

**Recomendación:**
- **Opción 1** = Mejor para ver interface gráfica
- **Opción 4** = Mejor para generar datos completos (10,000 tweets)

### 5c. Lo que sucede mientras corre Locust

**Locust:**
- Genera requests HTTP a `/api/tweets`
- Cada request contiene: municipio, temperatura, humedad, clima
- Monitorea: requests/segundo, latencia, errores

**Backend:**
- Rust API recibe y valida
- Envía a Go Service via gRPC
- Go envía a Kafka Y RabbitMQ simultáneamente
- Consumidores reciben y guardan en Valkey

---

## 📊 PASO 6: VER DATOS EN GRAFANA

Mientras Locust corre (o después), abre Grafana.

### 6a. Ejecutar script de Grafana

```bash
./run-grafana.sh

Selecciona opción 1: Abrir en navegador
```

### 6b. Acceder a Grafana

```
URL: http://localhost:3000
Usuario: admin
Contraseña: admin123
```

### 6c. Crear Dashboard

1. Click en `+` → `Dashboard`
2. Click en `Add Panel`
3. Selecciona tipo de gráfico (Bar, Stat, etc.)
4. Configura Data Source:
   - Si no existe "Redis", click en engranaje
   - `Configuration` → `Data Sources` → `Add data source`
   - Tipo: Redis
   - URL: `redis://valkey:6379`
   - Click `Save & test`
5. En el panel, en la sección "Queries", ejecuta:

```
GET weather:sunny
GET weather:cloudy
GET weather:rainy
GET weather:foggy
```

**Resultado:** Un gráfico mostrando contadores en tiempo real

---

## 💾 PASO 7: VERIFICAR DATOS EN VALKEY

Abre otra terminal y accede a la base de datos.

### 7a. Ejecutar script de Valkey

```bash
./run-valkey.sh

Selecciona opción 3: Ver contadores de clima
```

### 7b. Lo que verás

```
GET weather:sunny
(integer) 2450

GET weather:cloudy
(integer) 1890

GET weather:rainy
(integer) 3120

GET weather:foggy
(integer) 2540
```

**Estos números aumentan mientras Locust envía tweets.**

### 7c. Datos detallados (Opcional)

```bash
./run-valkey.sh
# Selecciona opción 4: Ver datos detallados

# Salida:
HGETALL weather:data:chinautla:sunny
1) "count"
2) "2450"
3) "avg_temp"
4) "24.5"
5) "avg_humidity"
6) "65.3"
```

---

## 📈 PASO 8: MONITOREAR ESCALADO (HPA)

El Horizontal Pod Autoscaler debe estar escalando el Rust API según carga.

```bash
# En otra terminal, ver HPA en vivo
kubectl get hpa -n weather-system -w

# Salida esperada:
NAME          REFERENCE           TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
rust-api-hpa  Deployment/rust-api 45%/30%   1         3         3          2m

# El REPLICAS debe cambiar de 1 → 2 → 3 según la carga
```

**¿Qué significa?**
- `TARGETS 45%/30%` = CPU está al 45%, threshold es 30%
- `REPLICAS 3` = Auto-escaló a 3 pods
- Si cae la carga, bajará a 1 pod

---

## 🔍 PASO 9: VER LOGS EN TIEMPO REAL

Para entender qué sucede en el backend:

```bash
# Logs del Go Service
kubectl logs -f deployment/go-deployment-1 -n weather-system

# Deberías ver algo como:
[2025-10-18 20:30:45] Recibido: municipio=chinautla, clima=sunny, temp=28°C, humidity=65%
[2025-10-18 20:30:45] Enviando a Kafka...
[2025-10-18 20:30:46] Enviando a RabbitMQ...
[2025-10-18 20:30:46] Almacenado en Valkey: weather:sunny = 2451
```

---

## 💡 PASO 10: COMPARAR KAFKA vs RabbitMQ (Opcional)

Puedes ver logs de ambos consumidores:

```bash
# Logs del Kafka Consumer
kubectl logs -f deployment/kafka-consumer -n weather-system

# Logs del RabbitMQ Consumer
kubectl logs -f deployment/rabbitmq-consumer -n weather-system

# Ambos deberían actualizar los mismos contadores en Valkey
# Kafka típicamente tiene más throughput
# RabbitMQ típicamente tiene menos latencia
```

---

## 📸 PASO 11: CAPTURAR EVIDENCIA

Para la calificación, captura:

1. **Screenshot de Grafana**
   - Dashboard mostrando contadores
   - En tiempo real mientras Locust corre

2. **Screenshot de Valkey**
   ```bash
   ./run-valkey.sh
   # Opción 3: Ver contadores
   ```

3. **Screenshot de HPA escalando**
   ```bash
   kubectl get hpa -n weather-system
   # Mostrar REPLICAS = 3
   ```

4. **Output de Locust**
   - Requests/segundo
   - Latencia promedio
   - Total de requests

5. **Logs de servicios**
   ```bash
   kubectl logs deployment/go-deployment-1 -n weather-system
   ```

---

## ✓ PASO 12: CHECKLIST FINAL

- [ ] ¿Se ejecutó `./deploy.sh` exitosamente?
- [ ] ¿Todos los pods están en estado `Running`?
- [ ] ¿Locust generó tweets (10,000)?
- [ ] ¿Grafana muestra datos actualizándose?
- [ ] ¿Valkey tiene contadores > 0?
- [ ] ¿HPA escaló a 3 replicas?
- [ ] ¿Capturaste screenshots de evidencia?
- [ ] ¿Documentaste métricas de desempeño?

---

## 🧹 LIMPIEZA (Solo si necesitas volver a empezar)

**⚠️ ADVERTENCIA:** Esto elimina TODO el namespace.

```bash
./cleanup.sh

# Confirma escribiendo "SI"
```

---

## 🆘 TROUBLESHOOTING

### Error: "No se pudo obtener la IP del Ingress"

```bash
# Esperar más tiempo
kubectl get ingress -n weather-system

# Si sigue sin IP, revisa eventos
kubectl get events -n weather-system --sort-by='.lastTimestamp'
```

### Error: "Imágenes no encontradas"

```bash
# Verificar que Zot esté corriendo
curl http://34.159.50.100:5000/v2/

# Ver imágenes disponibles
curl http://34.159.50.100:5000/v2/_catalog
```

### Error: "Locust no puede conectar"

```bash
# Verificar que la IP del Ingress es correcta
./run-tests.sh

# Probar conectividad manual
curl http://<INGRESS_IP>/health
```

### Error: "Valkey vacío"

```bash
# Normal si Locust no ha enviado datos
# Espera a que Locust termine de generar tweets
# Luego:
./run-valkey.sh
# Selecciona opción 3
```

---

## 📞 CONTACTO Y SOPORTE

**Carnet:** 202200129  
**Municipio:** Chinautla  
**Universidad:** USAC  
**Documento:** SO1 - Proyecto 3  

**Archivos relacionados:**
- `TECNICA.md` - Documentación técnica completa
- `RESUMEN.md` - Resumen ejecutivo
- `EJEMPLOS.md` - Ejemplos de comandos
- `SCRIPTS.md` - Descripción de scripts
- `README.md` - Quick start

---

*Última actualización: 18 Octubre 2025*
