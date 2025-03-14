from fastapi import FastAPI  # type: ignore
import os
import json
from typing import List
from models.models import logContainer
from models.models import logMemory
import matplotlib.pyplot as plt
import numpy as np
from matplotlib.dates import ConciseDateFormatter
# from matplotlib import pyplot as plt
# from matplotlib import numpy as np
# from matplotlib.dates import ConciseDateFormatter

app = FastAPI()


@app.get("/graph")
def get_graphs():
    ruta_memory = "./logs/memory.json"
    ruta_cont = "./logs/logs.json"

    try:
        with open(ruta_memory, 'r', encoding='utf-8') as file_memory:
            datos_memory = json.load(file_memory)

        with open(ruta_cont, 'r', encoding='utf-8') as file_cont:
            datos_cont = json.load(file_cont)

    except Exception as e:
        print(f"Error reading files: {e}")
        return {"error": "Failed to read log files"}

    # Grafica de uso de RAM
    x_memory = range(len(datos_memory))

    y_memory = [(elem.get('usage_ram_mb', 0) / elem.get('total_ram_mb', 1))
                * 100 for elem in datos_memory]

    plt.figure()
    plt.plot(x_memory, y_memory)
    plt.xlabel('Iteraciones')
    plt.ylabel('Porcentaje uso Memoria')
    plt.title('Uso de memoria')
    plt.savefig("./graphs/memoria.png")

    # Grafica de uso de CPU
    x_cont = range(len(datos_cont))
    y_cpu = [elem.get('cpu_usage', 0) for elem in datos_cont]

    plt.figure()
    plt.plot(x_cont, y_cpu)
    plt.xlabel('Iteraciones')
    plt.ylabel('Uso de CPU (%)')
    plt.title('Uso de CPU por contenedor')
    plt.savefig("./graphs/cpu.png")

    return {"graficas": "generadas"}


@app.post("/logs")
def get_logs(logs_proc: List[logContainer]):
    logs_file = 'logs/logs.json'

    # Depuración: Imprimir los datos recibidos
    print("Datos recibidos en /logs:")
    for log in logs_proc:
        print(log.dict())

    # Checamos si existe el archivo logs.json
    if os.path.exists(logs_file):
        # Leemos el archivo logs.json
        with open(logs_file, 'r') as file:
            existing_logs = json.load(file)
    else:
        # Sino existe, creamos una lista vacía
        existing_logs = []

    # Agregamos los nuevos logs a la lista existente
    new_logs = [log.dict() for log in logs_proc]
    existing_logs.extend(new_logs)

    # Escribimos la lista de logs en el archivo logs.json
    with open(logs_file, 'w') as file:
        json.dump(existing_logs, file, indent=4)

    return {"received": True}


@app.post("/memory")
def get_memory(logs_memory1: List[logMemory]):
    logs_file = 'logs/memory.json'

    # Depuración: Imprimir los datos recibidos
    print("Datos recibidos en /memory:")
    for log in logs_memory1:
        print(log.dict())

    # Checamos si existe el archivo logs.json
    if os.path.exists(logs_file):
        # Leemos el archivo logs.json
        with open(logs_file, 'r') as file:
            existing_logs = json.load(file)
    else:
        # Sino existe, creamos una lista vacía
        existing_logs = []

    # Agregamos los nuevos logs a la lista existente
    new_logs = [log.dict() for log in logs_memory1]
    existing_logs.extend(new_logs)

    # Escribimos la lista de logs en el archivo logs.json
    with open(logs_file, 'w') as file:
        json.dump(existing_logs, file, indent=4)

    return {"received": True}
