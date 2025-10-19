## Configuración de Grafana para el Dashboard

### Pasos para Crear el Dashboard

1. **Acceder a Grafana**
   - URL: `http://grafana.local`
   - Usuario: `admin`
   - Contraseña: `admin123`

2. **Crear Data Source (si no existe)**
   - Ir a: Configuration → Data Sources
   - Click: "Add data source"
   - Tipo: Redis
   - URL: `redis://valkey:6379`
   - Guardar y Probar

3. **Crear Dashboard**
   - Click: "+" → Dashboard
   - Click: "Add Panel"
   - Elegir: "Bar Chart"

4. **Configurar Panel**
   - **Título:** "Total de Reportes por Condición Climática"
   - **Query:**
     ```
     SELECT 
       'Sunny' as "Clima",
       CAST(GET weather:sunny AS INT) as "Reportes"
     UNION ALL
     SELECT 'Cloudy', CAST(GET weather:cloudy AS INT)
     UNION ALL
     SELECT 'Rainy', CAST(GET weather:rainy AS INT)
     UNION ALL
     SELECT 'Foggy', CAST(GET weather:foggy AS INT)
     ```

   - **O usar Redis queries directamente:**
     - `GET weather:sunny` → Etiqueta: "Sunny"
     - `GET weather:cloudy` → Etiqueta: "Cloudy"
     - `GET weather:rainy` → Etiqueta: "Rainy"
     - `GET weather:foggy` → Etiqueta: "Foggy"

5. **Opciones del Gráfico**
   - Panel Options:
     - X-axis: "Condición Climática"
     - Y-axis: "Número de Reportes"
   - Legend: Mostrar legenda

6. **Guardar Dashboard**
   - Nombre: "Clima Chinautla"
   - Click: Save

### Queries Redis Alternativas

Si la interfaz de Grafana no interpreta bien las queries, usar formato Redis:

```
# Sunny
GET weather:sunny

# Cloudy
GET weather:cloudy

# Rainy
GET weather:rainy

# Foggy
GET weather:foggy
```

Luego, en las opciones del panel, mapear estos valores a etiquetas.

### Panel Adicional: Datos Detallados por Municipio

```
HGETALL weather:data:chinautla:sunny
```

Esto retornará: `temperature` y `humidity` para cada condición.

### Alertas (Opcional)

Se pueden configurar alertas para:
- Cuando los reportes superen cierto umbral
- Cuando algún servicio no responda

Ir a: Alerting → Alert Rules → New Alert Rule

---

**Nota:** Para que el dashboard muestre datos, debe haber ejecutado Locust y procesado al menos 1 tweet.

