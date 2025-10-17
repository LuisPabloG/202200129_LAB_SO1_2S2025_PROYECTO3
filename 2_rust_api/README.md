# API REST en Rust

Este componente es una API REST desarrollada en Rust utilizando Actix Web. La API recibe peticiones de Locust y las envía al servicio Go mediante HTTP.

## Estructura del Proyecto

```
.
├── Cargo.toml
├── Dockerfile
├── src/
│   └── main.rs
└── k8s/
    └── deployment.yaml
```

## Modelo de Datos

La API maneja la siguiente estructura de datos para los tweets del clima:

```rust
#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweet {
    municipality: String,
    temperature: i32,
    humidity: i32,
    weather: String,
}
```

## Endpoints

- `POST /api/v1/tweet`: Recibe un tweet del clima y lo envía al servicio Go.

## Compilación y Ejecución Local

```bash
cargo build --release
./target/release/rust-api
```

## Compilación y Subida al Registry

```bash
docker build -t rust-api:latest .
docker tag rust-api:latest <IP-ZOT>:5000/202200129/rust-api:latest
docker push <IP-ZOT>:5000/202200129/rust-api:latest
```

## Despliegue en Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
```