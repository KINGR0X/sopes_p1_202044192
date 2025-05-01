import json
import sys
import random
from locust import HttpUser, task, constant
from locust.env import Environment
from locust.exception import RescheduleTask
from collections import defaultdict

# Configuración
TOTAL_REQUESTS = 10000
CONCURRENT_USERS = 10
REQUEST_COUNT = 0
STATS = defaultdict(int)  # Para contar por tipo de clima

# countries = ["GT", "SV", "NI", "CR", "PA"]
# weather_options = [0, 1, 2]  # rainy=0, cloudy=1, sunny=2


class LoadTester(HttpUser):
    wait_time = constant(0.1)
    test_data = []
    host = ""

    def on_start(self):
        """Cargar datos una sola vez"""
        if not self.test_data:
            self.load_test_data()

    def load_test_data(self):
        """Cargar datos desde input.json"""
        try:
            with open('input.json', 'r') as f:
                self.test_data = json.load(f)
            print(f"\nDatos cargados: {len(self.test_data)} registros")
        except Exception as e:
            print(f"\nERROR cargando datos: {str(e)}")
            sys.exit(1)

    @task
    def send_data(self):
        global REQUEST_COUNT

        if REQUEST_COUNT >= TOTAL_REQUESTS:
            self.environment.runner.quit()
            return

        # Seleccionar dato aleatorio
        data = random.choice(self.test_data)
        weather_type = data["weather"]
        STATS[weather_type] += 1
        REQUEST_COUNT += 1

        try:
            with self.client.post("/input",
                                  json=data,
                                  catch_response=True) as response:

                if response.status_code == 200:
                    print(
                        f"SI [{REQUEST_COUNT}/{TOTAL_REQUESTS}] W{weather_type} {data['country']}")
                    response.success()
                else:
                    print(
                        f"X [{REQUEST_COUNT}/{TOTAL_REQUESTS}] W{weather_type} {data['country']}: {response.status_code}")
                    response.failure(f"Código {response.status_code}")
                    raise RescheduleTask()

        except Exception as e:
            print(f"ERROR [{REQUEST_COUNT}/{TOTAL_REQUESTS}] Error: {str(e)}")
            raise RescheduleTask()


def print_stats():
    """Imprimir estadísticas detalladas"""
    print("\n" + "="*50)
    print("ESTADÍSTICAS FINALES")
    print("="*50)

    print(f"\nTotal peticiones enviadas: {REQUEST_COUNT}")

    print("\nDistribución por tipo de clima:")
    weather_names = {0: "Lluvioso", 1: "Nublado", 2: "Soleado"}
    for wt, count in sorted(STATS.items()):
        print(f"- {weather_names.get(wt, f'Desconocido ({wt})')
                   }: {count} ({count/REQUEST_COUNT:.1%})")

    print("\nDistribución por posición en el array:")
    for i in range(len(LoadTester.test_data)):
        count = STATS[i] if i in STATS else 0
        print(f"- Posición {i}: {count} ({count/REQUEST_COUNT:.1%})")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("\nUso: python app.py http://host-base")
        print("Ejemplo: python app.py http://34.94.158.179.nip.io")
        sys.exit(1)

    host = sys.argv[1]
    print(f"\nIniciando prueba de carga:")
    print(f"- Usuarios: {CONCURRENT_USERS}")
    print(f"- Peticiones: {TOTAL_REQUESTS}")
    print(f"- Target: {host}")

    env = Environment(user_classes=[LoadTester])
    env.host = host
    runner = env.create_local_runner()

    try:
        runner.start(CONCURRENT_USERS, spawn_rate=CONCURRENT_USERS)
        runner.greenlet.join()
    except KeyboardInterrupt:
        print("\nPrueba interrumpida manualmente")
    finally:
        print_stats()
        runner.quit()
