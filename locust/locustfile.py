from locust import HttpUser, task, between
import random

class WeatherTweetUser(HttpUser):
    wait_time = between(1, 5)

    @task
    def send_tweet(self):
        municipalities = ["mixco", "guatemala", "amatitlan", "chinautla"]
        weathers = ["sunny", "cloudy", "rainy", "foggy"]
        
        # Based on your carnet 202200129, the municipality is chinautla
        municipality = "chinautla"

        self.client.post("/tweet", json={
            "municipality": municipality,
            "temperature": random.randint(-10, 40),
            "humidity": random.randint(0, 100),
            "weather": random.choice(weathers)
        })
