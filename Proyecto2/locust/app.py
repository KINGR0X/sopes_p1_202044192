import random
import json
import signal
from locust import HttpUser, task, between
from datetime import datetime


class RustLoadTester(HttpUser):
    wait_time = between(1, 3)
    # Guate, El Salvador, Nicaragua, Costa Rica, Panama
    countries = ["GT", "SV", "NI", "CR", "PA"]
    weather_options = [0, 1, 2]  # rainy=0, cloudy=1, sunny=2
    sent_data = []
    filename = f"sent_data_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # Registrar manejador para señales de interrupción
        signal.signal(signal.SIGINT, self.save_data_handler)
        signal.signal(signal.SIGTERM, self.save_data_handler)

    def save_data_handler(self, signum, frame):
        """Manejador para señales de interrupción que guarda los datos antes de salir"""
        print(
            f"\nRecibida señal {signum}, guardando datos en {self.filename}...")
        self.save_data_to_file()
        exit(0)

    def save_data_to_file(self):
        """Guarda los datos enviados en un archivo JSON"""
        with open(self.filename, 'w') as f:
            json.dump(self.sent_data, f, indent=2)
        print(f"Datos guardados en {self.filename}")

    @task
    def send_tweet(self):
        data = {
            "descripcion": self.generate_description(),
            "country": random.choice(self.countries),
            "weather": random.choice(self.weather_options)
        }

        # Guardar los datos antes de enviarlos
        self.sent_data.append(data)

        self.client.post("/input", json=data)

    def generate_description(self):
        weather = random.choice(["rainy", "cloudy", "sunny"])
        return f"It's {weather} today"

    def on_stop(self):
        """Se ejecuta cuando el test termina normalmente"""
        self.save_data_to_file()
