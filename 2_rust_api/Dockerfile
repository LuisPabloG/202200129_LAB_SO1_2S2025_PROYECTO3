FROM rust:1.73 as builder
WORKDIR /usr/src/app
COPY . .
RUN cargo build --release

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/target/release/weather-api /usr/local/bin/weather-api

# Variables de entorno
ENV PORT=8000
ENV GO_SERVICE_URL=http://go-service:8080/api/v1/tweet

# Metadatos
LABEL org.opencontainers.image.authors="Luis Pablo Garcia - 202200129"
LABEL org.opencontainers.image.description="API REST en Rust para el Proyecto 3 de Sistemas Operativos 1"

EXPOSE 8000
CMD ["weather-api"]