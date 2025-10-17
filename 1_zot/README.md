# Configuración de Zot Registry

## Requisitos
- Una VM en GCP con Debian/Ubuntu
- Puertos 5000 abierto en el firewall

## Pasos para Configurar Zot

1. Crear una VM en GCP:
```bash
gcloud compute instances create zot-registry-202200129 \
  --project=proyecto-3-475405 \
  --zone=us-central1-a \
  --machine-type=e2-small \
  --image-family=debian-11 \
  --image-project=debian-cloud \
  --boot-disk-size=10GB \
  --tags=http-server,https-server
```

2. Configurar reglas de firewall:
```bash
gcloud compute firewall-rules create allow-zot-registry \
  --project=proyecto-3-475405 \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:5000 \
  --source-ranges=0.0.0.0/0
```

3. Conectarse a la VM:
```bash
gcloud compute ssh zot-registry-202200129 --project=proyecto-3-475405 --zone=us-central1-a
```

4. Ejecutar el script setup_zot.sh:
```bash
# Primero copia el script a la VM
gcloud compute scp setup_zot.sh zot-registry-202200129:~/ --project=proyecto-3-475405 --zone=us-central1-a

# Conéctate a la VM y ejecuta
gcloud compute ssh zot-registry-202200129 --project=proyecto-3-475405 --zone=us-central1-a
chmod +x ~/setup_zot.sh
sudo ./setup_zot.sh
```

5. Verifica que Zot esté funcionando:
```bash
curl http://<IP-DE-LA-VM>:5000/v2/
```

## Uso de Zot Registry

1. Para autenticarse desde Docker:
```bash
docker login <IP-DE-LA-VM>:5000
```

2. Para etiquetar y subir una imagen:
```bash
docker tag mi-imagen:latest <IP-DE-LA-VM>:5000/202200129/mi-imagen:latest
docker push <IP-DE-LA-VM>:5000/202200129/mi-imagen:latest
```

3. Para descargar una imagen:
```bash
docker pull <IP-DE-LA-VM>:5000/202200129/mi-imagen:latest
```

4. Para acceder a la interfaz web de Zot:
Visita `http://<IP-DE-LA-VM>:5000` en tu navegador.