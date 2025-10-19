from locust import HttpUser, task, between
import json
import random

# Municipios asignados por carnet: 202200129 (termina en 9) = Chinautla
MUNICIPALITIES = {
    0: "mixco",
    1: "mixco", 
    2: "mixco",
    3: "guatemala",
    4: "guatemala",
    5: "guatemala",
    6: "amatitlan",
    7: "amatitlan",
    8: "chinautla",
    9: "chinautla"
}

WEATHERS = ["sunny", "cloudy", "rainy", "foggy"]

class WeatherTweetUser(HttpUser):
    wait_time = between(0.1, 0.5)  # Espera entre 100ms y 500ms entre requests

    @task
    def send_weather_tweet(self):
        """Envía un tweet de clima"""
        # Obtener municipio según carnet (202200129 -> 9 -> chinautla)
        municipality = "chinautla"
        
        # Generar datos aleatorios
        temperature = random.randint(15, 35)  # Temperatura en Celsius
        humidity = random.randint(30, 90)      # Humedad en porcentaje
        weather = random.choice(WEATHERS)      # Clima aleatorio
        
        # Crear payload
        payload = {
            "municipality": municipality,
            "temperature": temperature,
            "humidity": humidity,
            "weather": weather
        }
        
        # Enviar POST request
        response = self.client.post(
            "/api/tweets",
            json=payload,
            headers={"Content-Type": "application/json"}
        )
        
        if response.status_code != 200:
            print(f"Error: {response.status_code} - {response.text}")

    @task
    def health_check(self):
        """Verifica el estado del servicio"""
        self.client.get("/health")
