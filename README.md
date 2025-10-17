<div align="center">

# **Universidad de San Carlos de Guatemala**
### **Facultad de Ingeniería – Escuela de Ciencias y Sistemas**
# PROYECTO 3 - WEATHER TWEETS SYSTEM

| | |
| :--- | :--- |
| **Curso:** | Laboratorio sistemas operativos 1 |
| **Estudiante:** | Luis Pablo Manuel García López |
| **Carnet:** | 202200129 |
| **Fecha:** | 22/10/2025 |

</div>

## Arquitectura del Sistema

El sistema implementa una arquitectura de microservicios distribuida para el procesamiento de "tweets" meteorológicos del municipio de Chinautla. La arquitectura está compuesta por los siguientes componentes:

### Componentes principales

1. **Zot Registry**: Registro de contenedores OCI privado para almacenar las imágenes Docker del proyecto.
2. **Rust API**: Punto de entrada REST que recibe los tweets meteorológicos.
3. **Go API**: Servicio intermediario que recibe solicitudes del API Rust y las envía a los servicios de escritura.
4. **Kafka Writer**: Servicio gRPC que publica mensajes en Kafka.
5. **RabbitMQ Writer**: Servicio gRPC que publica mensajes en RabbitMQ.
6. **Kafka Consumer**: Consumidor que procesa mensajes de Kafka y los almacena en Valkey.
7. **RabbitMQ Consumer**: Consumidor que procesa mensajes de RabbitMQ y los almacena en Valkey.
8. **Valkey**: Base de datos en memoria compatible con Redis para almacenar los datos procesados.
9. **Prometheus**: Sistema de monitoreo para recopilar métricas.
10. **Grafana**: Herramienta para visualizar las métricas y datos meteorológicos.
11. **Locust**: Herramienta para realizar pruebas de carga.

## Estructura del Proyecto

```
.
├── 1_zot/               # Configuración del registro de contenedores Zot
├── 2_rust_api/          # API REST en Rust (Actix-web)
├── 3_go_api/            # API en Go (HTTP y gRPC client)
├── 4_go_kafka_writer/   # Servicio gRPC en Go para escritura en Kafka
├── 5_go_rabbit_writer/  # Servicio gRPC en Go para escritura en RabbitMQ
├── 6_go_kafka_consumer/ # Consumidor de Kafka en Go
├── 7_go_rabbit_consumer/ # Consumidor de RabbitMQ en Go
├── k8s/                 # Configuraciones de Kubernetes
├── create_oci_images.sh # Script para crear imágenes OCI
├── deploy_kubernetes.sh # Script para desplegar en Kubernetes
└── README.md            # Este archivo
```

## Cómo ejecutar el proyecto

### Prerrequisitos

- Docker
- kubectl
- Cuenta de Google Cloud con acceso a GKE
- gcloud CLI configurado
- Cluster de Kubernetes creado en GCP (gke-sopes3-202200129)

### Pasos para la ejecución

1. **Configurar el registro Zot**

```bash
cd 1_zot
./setup_zot.sh
```

2. **Crear imágenes OCI y publicarlas en Zot**

```bash
./create_oci_images.sh
```

3. **Desplegar en Kubernetes (GKE)**

```bash
./deploy_kubernetes.sh
```

4. **Acceder a los servicios**

Una vez desplegado, podrá acceder a:

- **Grafana**: http://<INGRESS_IP>/grafana (usuario: admin, contraseña: admin)
- **Locust**: http://<INGRESS_IP>/locust
- **Enviar tweets**: Envíe POST a http://<INGRESS_IP>/tweet con el formato:

```json
{
  "municipality": "chinautla",
  "temperature": 25.5,
  "humidity": 60.2,
  "timestamp": 1629843600
}
```

## Dashboard de Chinautla

El dashboard de Grafana muestra la siguiente información para el municipio de Chinautla:

- Gráfico de temperatura a lo largo del tiempo
- Gráfico de humedad a lo largo del tiempo
- Indicadores de temperatura y humedad actual
- Distribución de mensajes entre Kafka y RabbitMQ
- Métrica de solicitudes por segundo
