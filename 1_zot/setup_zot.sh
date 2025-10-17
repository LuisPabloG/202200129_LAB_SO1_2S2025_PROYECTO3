#!/bin/bash

# Instalar Docker
sudo apt-get update
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Añadir la clave GPG oficial de Docker
sudo mkdir -m 0755 -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Configurar el repositorio de Docker
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Instalar Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Crear directorio para Zot
sudo mkdir -p /opt/zot/config
sudo mkdir -p /opt/zot/data

# Crear configuración de Zot
cat << EOF | sudo tee /opt/zot/config/config.json
{
  "distSpecVersion": "1.1.0",
  "storage": {
    "rootDirectory": "/var/lib/zot/",
    "dedupe": true,
    "gc": true,
    "gcDelay": 1,
    "gcInterval": 1
  },
  "http": {
    "address": "0.0.0.0",
    "port": "5000"
  },
  "log": {
    "level": "debug"
  },
  "extensions": {
    "search": {
      "enable": true
    },
    "sync": {
      "enable": true
    }
  }
}
EOF

# Ejecutar Zot como un contenedor Docker
sudo docker run -d \
  --name zot \
  --restart always \
  -p 5000:5000 \
  -v /opt/zot/config:/etc/zot \
  -v /opt/zot/data:/var/lib/zot \
  ghcr.io/project-zot/zot-linux-amd64:latest

echo "Zot Registry está ejecutándose en el puerto 5000"