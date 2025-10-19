# Use the official Rust image.
# https://hub.docker.com/_/rust
FROM rust:1.78.0 as builder

# Create a new empty shell project
WORKDIR /usr/src/app
COPY . .

# Build the project
RUN cargo build --release

# Production image
FROM debian:buster-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/target/release/api-rust .
CMD ["./api-rust"]
