# Build stage
FROM rust:latest as builder

WORKDIR /app

COPY Cargo.toml Cargo.toml
COPY build.rs build.rs
COPY src src
COPY ../proto ../proto

RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/weather_api /usr/local/bin/

EXPOSE 8080

CMD ["weather_api"]
