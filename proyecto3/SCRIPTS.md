# 🚀 SCRIPTS DE EJECUCIÓN - Proyecto 3 SOPES

**Carnet:** 202200129 | **Municipio:** Chinautla | **Año:** 2025

---

## 📋 INICIO RÁPIDO

```bash
# Paso 0: Ver guía completa
./INICIO.sh

# Paso 1: Actualizar registry (reemplaza IP)
./update-registry.sh 34.159.50.100:5000

# Paso 2: Compilar imágenes Docker
./build-images.sh 34.159.50.100:5000

# Paso 3: Desplegar infraestructura (15-20 min)
./deploy.sh

# Paso 4: Monitorear (en otra terminal)
kubectl get pods -n weather-system -w
```

---

## 📜 DESCRIPCIÓN DE SCRIPTS

### 🟢 **INICIO.sh** - Guía Interactiva de Inicio
- **Función:** Muestra guía paso a paso con instrucciones
- **Uso:** `./INICIO.sh`
- **Salida:** Guía visual con todos los pasos
- **Tiempo:** < 1 segundo

### 🟢 **run-locust.sh** - Generador de Carga
- **Función:** Ejecuta pruebas de carga con Locust
- **Opciones:**
  - 1: Modo Web (Interface gráfica) → http://localhost:8089
  - 2: Ligero (1,000 tweets, 5 usuarios)
  - 3: Medio (5,000 tweets, 10 usuarios)
  - 4: Pesado (10,000 tweets, 20 usuarios)
  - 5: Personalizado
- **Uso:** `./run-locust.sh`
- **Tiempo:** Variable (según opción)

```bash
# Ejemplo: Generar 10,000 tweets
./run-locust.sh
# Selecciona opción 4
```

### 🟢 **run-grafana.sh** - Acceso a Dashboard
- **Función:** Accede a Grafana para visualizar datos
- **Opciones:**
  - 1: Abrir en navegador
  - 2: Ver URL
  - 3: Status del pod
  - 4: Ver logs
- **Uso:** `./run-grafana.sh`
- **URL:** http://localhost:3000
- **Credenciales:** admin / admin123

```bash
# Ejemplo: Acceder a Grafana
./run-grafana.sh
# Selecciona opción 1
```

### 🟢 **run-valkey.sh** - Acceso a Base de Datos
- **Función:** Conecta a Valkey (Redis) para queries
- **Opciones:**
  - 1: CLI interactivo
  - 2: Ver todas las claves
  - 3: Ver contadores de clima
  - 4: Ver datos detallados
  - 5: Estadísticas completas
  - 6: Monitor en tiempo real
  - 7: Limpiar datos (peligro)
- **Uso:** `./run-valkey.sh`

```bash
# Ejemplo: Ver contadores en tiempo real
./run-valkey.sh
# Selecciona opción 3
```

### 🟢 **run-tests.sh** - Status General
- **Función:** Verifica status de cluster y muestra opciones
- **Uso:** `./run-tests.sh`
- **Salida:** Status de pods, Ingress IP, comandos útiles

### 🟡 **deploy.sh** - Desplegar Infraestructura
- **Función:** Desplega Kubernetes manifests en orden
- **Secuencia:**
  1. Crea namespace
  2. Valkey (5 min)
  3. Kafka (8 min)
  4. RabbitMQ (5 min)
  5. Servicios (Go, Rust)
  6. Grafana
  7. Ingress
- **Uso:** `./deploy.sh`
- **Tiempo:** 15-20 minutos
- **Requisito:** Imágenes Docker en Zot Registry

### 🟡 **build-images.sh** - Compilar Imágenes Docker
- **Función:** Compila 4 imágenes y las envía a Zot
- **Imágenes:**
  - weather-api-rust
  - go-services-deployment1
  - go-kafka-consumer
  - go-rabbitmq-consumer
- **Uso:** `./build-images.sh ZOT_IP:PORT`
- **Ejemplo:** `./build-images.sh 34.159.50.100:5000`
- **Tiempo:** 5-10 minutos
- **Requisito:** Zot Registry corriendo

### 🟡 **update-registry.sh** - Actualizar IP de Zot
- **Función:** Actualiza IP de Zot Registry en YAML files
- **Uso:** `./update-registry.sh ZOT_IP:PORT`
- **Ejemplo:** `./update-registry.sh 34.159.50.100:5000`
- **Modifica:** Todos los k8s-manifests/*.yaml

### 🔵 **monitor.sh** - Monitoreo Continuo
- **Función:** Monitorea en vivo los pods y HPA
- **Uso:** `./monitor.sh`
- **Salida:** Actualización cada 2 segundos

### 🔵 **cleanup.sh** - Limpiar Namespace
- **Función:** Elimina completamente el namespace weather-system
- **Uso:** `./cleanup.sh`
- **Advertencia:** ⚠️ Borra TODOS los datos

### 🔵 **show-info.sh** - Información Completa
- **Función:** Muestra resumen visual de todo el proyecto
- **Uso:** `./show-info.sh`
- **Salida:** Arquitectura, archivos, checklist, etc.

### 🔵 **make-executable.sh** - Permisos
- **Función:** Hace todos los scripts ejecutables
- **Uso:** `./make-executable.sh`

---

## 🔄 FLUJO COMPLETO DE EJECUCIÓN

```
┌─────────────────────────────────────────────────────────────┐
│ PASO 0: Verificar conexión con cluster                     │
│ $ kubectl get nodes                                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 1: Actualizar Zot Registry                            │
│ $ ./update-registry.sh 34.159.50.100:5000                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 2: Compilar imágenes Docker (5-10 min)               │
│ $ ./build-images.sh 34.159.50.100:5000                     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 3: Desplegar infraestructura (15-20 min)             │
│ $ ./deploy.sh                                               │
│ (En otra terminal: $ kubectl get pods -n weather-system -w)│
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 4: Ejecutar Locust (Pruebas de carga)                │
│ $ ./run-locust.sh                                           │
│ Opción 4: Pesado (10,000 tweets)                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 5: Ver datos en Grafana (En otra terminal)           │
│ $ ./run-grafana.sh                                          │
│ Opción 1: Abrir en navegador                               │
│ URL: http://localhost:3000                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ PASO 6: Verificar datos en Valkey (En otra terminal)      │
│ $ ./run-valkey.sh                                           │
│ Opción 3: Ver contadores                                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
            ✓ SISTEMA COMPLETAMENTE FUNCIONAL
```

---

## 🎯 CASOS DE USO COMUNES

### Ejecutar 10,000 tweets automáticos
```bash
./run-locust.sh
# Selecciona opción 4
```

### Ver contadores en Valkey mientras Locust corre
```bash
./run-valkey.sh
# Selecciona opción 3
# Actualiza cada segundos con valores reales
```

### Abrir Grafana y crear dashboard
```bash
./run-grafana.sh
# Selecciona opción 1
# Luego crea panel con queries de Redis
```

### Monitorear HPA escalando
```bash
kubectl get hpa -n weather-system -w
```

### Ver logs del Go Service en tiempo real
```bash
kubectl logs -f deployment/go-deployment-1 -n weather-system
```

### Limpiar TODO (solo antes de nuevo despliegue)
```bash
./cleanup.sh
```

---

## ⚠️ ADVERTENCIAS IMPORTANTES

### ANTES DE DESPLEGAR
- [ ] Tienes `kubectl` configurado con credenciales GKE
- [ ] IP de Zot Registry actualizada
- [ ] Zot Registry está corriendo y accesible
- [ ] Cluster tiene suficiente recursos (3+ nodos o n1-standard-2+)

### DURANTE LAS PRUEBAS
- [ ] NO destruyas el cluster (debe estar 24+ horas)
- [ ] NO limpies Valkey antes de capturar pantallazos
- [ ] Mantén HPA habilitado (mostrará escalado)
- [ ] Documenta métricas de antes/después

### DESPUÉS DE PRUEBAS
- [ ] Captura screenshots de Grafana
- [ ] Exporta datos de Valkey
- [ ] Guarda logs importantes
- [ ] Actualiza documentación con resultados

---

## 🆘 TROUBLESHOOTING

### "No hay conexión con el cluster"
```bash
# Recargar credenciales
gcloud container clusters get-credentials proyecto3-sopes-1 \
  --zone us-central1-c --project proyecto-3-475405
```

### "Imágenes no encontradas en Zot"
```bash
# Verificar que Zot Registry está corriendo
curl http://34.159.50.100:5000/v2/

# Ver imágenes disponibles
curl http://34.159.50.100:5000/v2/_catalog
```

### "Pods en estado Pending"
```bash
# Ver evento del pod
kubectl describe pod <POD_NAME> -n weather-system

# Ver eventos del namespace
kubectl get events -n weather-system --sort-by='.lastTimestamp'
```

### "Locust no se conecta a API"
```bash
# Verificar Ingress IP
kubectl get ingress -n weather-system

# Probar conectividad
curl http://<INGRESS_IP>/health
```

---

## 📊 ARCHIVOS RELACIONADOS

| Archivo | Descripción |
|---------|-------------|
| `deploy.sh` | Desplegar infraestructura |
| `build-images.sh` | Compilar imágenes Docker |
| `run-locust.sh` | Ejecutar pruebas de carga |
| `run-grafana.sh` | Acceder a dashboard |
| `run-valkey.sh` | Acceder a base de datos |
| `run-tests.sh` | Ver status general |
| `TECNICA.md` | Documentación técnica (70+ p) |
| `RESUMEN.md` | Resumen ejecutivo |
| `README.md` | Quick start |

---

## 📞 SOPORTE

Para más información, consulta:
- `TECNICA.md` - Documentación técnica completa
- `EJEMPLOS.md` - Ejemplos de comandos
- `README.md` - Quick start guide

**Carnet:** 202200129  
**Municipio:** Chinautla  
**Universidad:** USAC  
**Año:** 2025

---

*Última actualización: 18 Octubre 2025*
